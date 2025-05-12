package syncer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

	// Logger is used to log messages
	Logger *slog.Logger
}

func (c *Config) Validate() error {
	if c.PollInterval <= 0 {
		return errors.New("poll interval must be positive")
	}
	if c.AcknowledgmentQueue == "" {
		return errors.New("acknowledgment queue name is required")
	}
	if c.Registry == nil {
		return errors.New("registry is required")
	}
	if c.Publisher == nil {
		return errors.New("publisher is required")
	}
	if c.Consumer == nil {
		return errors.New("consumer is required")
	}
	if c.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}

type Syncer struct {
	ctx                 context.Context
	registry            repository.Registry
	publisher           messaging.Publisher
	consumer            messaging.Consumer
	logger              *slog.Logger
	pollInterval        time.Duration
	acknowledgmentQueue string
}

func New(ctx context.Context, config Config) (*Syncer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Syncer{
		ctx:                 ctx,
		registry:            config.Registry,
		publisher:           config.Publisher,
		consumer:            config.Consumer,
		logger:              config.Logger,
		pollInterval:        config.PollInterval,
		acknowledgmentQueue: config.AcknowledgmentQueue,
	}, nil
}

func (s *Syncer) Start() error {
	go s.startChangePoller(s.ctx)

	if err := s.consumer.Consume(s.ctx, s.acknowledgmentQueue, s.handleAcknowledgment); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	<-s.ctx.Done()

	return s.ctx.Err()
}

func (s *Syncer) startChangePoller(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.processPendingChanges(ctx); err != nil {
				s.logger.ErrorContext(ctx, "Error processing pending changes", "error", err)
			}
		}
	}
}

func (s *Syncer) processPendingChanges(ctx context.Context) error {
	var err error
	changes, err := s.registry.GetChangeRepository().GetChangesWithPendingTargets(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending changes: %w", err)
	}

	for _, change := range changes {
		if err = s.processChange(ctx, change); err != nil {
			s.logger.ErrorContext(ctx, "Error processing change", "error", err, "change_id", change.ID)
		}
	}

	return nil
}

func (s *Syncer) processChange(ctx context.Context, change *beacondomain.Change) error {
	var err error
	targets, err := s.registry.GetChangeRepository().GetPendingTargetsForChange(ctx, change.ID)
	if err != nil {
		return fmt.Errorf("failed to get pending targets for change %s: %w", change.ID, err)
	}

	for _, target := range targets {
		if err = s.sendChangeToTarget(ctx, change, target); err != nil {
			s.logger.ErrorContext(
				ctx,
				"Error sending change",
				"error", err,
				"change_id", change.ID,
				"target", target.Server.HostName,
			)
		}
	}

	return nil
}

func (s *Syncer) sendChangeToTarget(
	ctx context.Context,
	change *beacondomain.Change,
	target controllerdomain.ChangeTarget,
) error {
	headers := messaging.Headers{
		messaging.HeaderKeyHost:    target.Server.HostName,
		messaging.HeaderKeyReplyTo: s.acknowledgmentQueue,
	}

	protoChange := convert.DomainChangeToProto(change)
	protoChangeBytes, err := proto.Marshal(protoChange)
	if err != nil {
		return fmt.Errorf("failed to marshal change: %w", err)
	}

	routingKey := fmt.Sprintf("server.%s.%s", target.Server.Type, target.Server.HostName)

	err = s.publisher.Publish(
		ctx,
		routingKey,
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

	host, ok := headers.GetString(messaging.HeaderKeyHost)
	if !ok {
		return messaging.NewConsumerError(errors.New("host header not found"), false)
	}

	changeID, err := uuid.Parse(ackMsg.GetChangeId())
	if err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to parse change_id: %w", err), false)
	}

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
