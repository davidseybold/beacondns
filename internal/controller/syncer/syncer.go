package syncer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	controllerdomain "github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/controller/repository"
	"github.com/davidseybold/beacondns/internal/convert"
	beacondomain "github.com/davidseybold/beacondns/internal/domain"
	beacondnspb "github.com/davidseybold/beacondns/internal/libs/gen/proto/beacondns/v1"
	"github.com/davidseybold/beacondns/internal/libs/messaging"
)

type Config struct {
	// PollInterval is how often to check for new changes
	PollInterval time.Duration

	// AcknowledgmentQueue is the queue to listen for acknowledgments
	AcknowledgmentQueue string

	// Registry provides access to repositories
	Registry repository.Registry

	// Publisher is used to send changes to servers
	Publisher messaging.Publisher

	// Consumer is used to receive acknowledgments
	Consumer messaging.Consumer
}

// Validate checks if the config is valid
func (c *Config) Validate() error {
	if c.PollInterval <= 0 {
		return fmt.Errorf("poll interval must be positive")
	}
	if c.AcknowledgmentQueue == "" {
		return fmt.Errorf("acknowledgment queue name is required")
	}
	if c.Registry == nil {
		return fmt.Errorf("registry is required")
	}
	if c.Publisher == nil {
		return fmt.Errorf("publisher is required")
	}
	if c.Consumer == nil {
		return fmt.Errorf("consumer is required")
	}
	return nil
}

// Syncer is responsible for syncing changes to servers and handling acknowledgments
type Syncer struct {
	ctx       context.Context
	registry  repository.Registry
	publisher messaging.Publisher
	consumer  messaging.Consumer

	pollInterval        time.Duration
	acknowledgmentQueue string
}

// New creates a new syncer
func New(ctx context.Context, config Config) (*Syncer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Syncer{
		ctx:                 ctx,
		registry:            config.Registry,
		publisher:           config.Publisher,
		consumer:            config.Consumer,
		pollInterval:        config.PollInterval,
		acknowledgmentQueue: config.AcknowledgmentQueue,
	}, nil
}

// Start begins the syncer's work
func (s *Syncer) Start() error {
	// Start the change poller
	go s.startChangePoller(s.ctx)

	// Start the acknowledgment listener
	if err := s.consumer.Consume(s.ctx, s.acknowledgmentQueue, s.handleAcknowledgment); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	<-s.ctx.Done()

	return s.ctx.Err()
}

// startChangePoller periodically checks for pending changes
func (s *Syncer) startChangePoller(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.processPendingChanges(ctx); err != nil {
				fmt.Printf("Error processing pending changes: %v\n", err)
			}
		}
	}
}

func (s *Syncer) processPendingChanges(ctx context.Context) error {
	changes, err := s.registry.GetChangeRepository().GetChangesWithPendingTargets(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending changes: %w", err)
	}

	for _, change := range changes {
		if err := s.processChange(ctx, change); err != nil {
			fmt.Printf("Error processing change %s: %v\n", change.ID, err)
		}
	}

	return nil
}

// processChange handles a single change and its targets
func (s *Syncer) processChange(ctx context.Context, change *beacondomain.Change) error {
	targets, err := s.registry.GetChangeRepository().GetPendingTargetsForChange(ctx, change.ID)
	if err != nil {
		return fmt.Errorf("failed to get pending targets for change %s: %w", change.ID, err)
	}

	for _, target := range targets {
		if err := s.sendChangeToTarget(ctx, change, target); err != nil {
			fmt.Printf("Error sending change %s to target %s: %v\n", change.ID, target.Server.HostName, err)
		}
	}

	return nil
}

func (s *Syncer) sendChangeToTarget(ctx context.Context, change *beacondomain.Change, target controllerdomain.ChangeTarget) error {
	headers := messaging.Headers{
		messaging.HeaderKeyHost:    target.Server.HostName,
		messaging.HeaderKeyReplyTo: s.acknowledgmentQueue,
	}

	protoChange := convert.DomainChangeToProto(change)
	protoChangeBytes, err := proto.Marshal(protoChange)
	if err != nil {
		return fmt.Errorf("failed to marshal change: %w", err)
	}

	err = s.publisher.Publish(
		ctx,
		fmt.Sprintf("server.%s.%s", target.Server.Type, target.Server.HostName),
		headers,
		protoChangeBytes,
	)
	if err != nil {
		return fmt.Errorf("failed to publish change: %w", err)
	}

	err = s.registry.GetChangeRepository().UpdateChangeTargetStatus(
		ctx,
		change.ID,
		target.Server.HostName,
		controllerdomain.ChangeTargetStatusSent,
	)
	if err != nil {
		return fmt.Errorf("failed to update target status: %w", err)
	}

	return nil
}

func (s *Syncer) handleAcknowledgment(body []byte, headers messaging.Headers) error {
	ackMsg := &beacondnspb.ChangeAck{}
	if err := proto.Unmarshal(body, ackMsg); err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to unmarshal acknowledgment: %w", err), false)
	}

	hostRaw, ok := headers[messaging.HeaderKeyHost]
	if !ok {
		return messaging.NewConsumerError(errors.New("missing server_id in message headers"), false)
	}

	host, ok := hostRaw.(string)
	if !ok {
		return messaging.NewConsumerError(errors.New("server_id is not a string"), false)
	}

	changeID, err := uuid.Parse(ackMsg.ChangeId)
	if err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to parse change_id: %w", err), false)
	}

	// Update the target status to INSYNC
	err = s.registry.GetChangeRepository().UpdateChangeTargetStatus(
		context.Background(),
		changeID,
		host,
		controllerdomain.ChangeTargetStatusInSync,
	)
	if err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to update target status: %w", err), true)
	}

	return nil
}
