package beacon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/proto"

	"github.com/davidseybold/beacondns/internal/convert"
	beacondnspb "github.com/davidseybold/beacondns/internal/gen/proto/beacondns/v1"
	"github.com/davidseybold/beacondns/internal/logger"
	"github.com/davidseybold/beacondns/internal/messaging"
	"github.com/davidseybold/beacondns/internal/model"
)

type ChangeListenerConfig struct {
	Consumer     messaging.Consumer
	Publisher    messaging.Publisher
	ChangeQueue  string
	Logger       *slog.Logger
	OnZoneChange func(ch *model.ZoneChange) error
}

type ChangeListener struct {
	logger       *slog.Logger
	consumer     messaging.Consumer
	publisher    messaging.Publisher
	changeQueue  string
	onZoneChange func(ch *model.ZoneChange) error
}

func NewChangeListener(config ChangeListenerConfig) *ChangeListener {
	if config.Logger == nil {
		config.Logger = logger.NewDiscardLogger()
	}

	return &ChangeListener{
		consumer:     config.Consumer,
		publisher:    config.Publisher,
		changeQueue:  config.ChangeQueue,
		logger:       config.Logger,
		onZoneChange: config.OnZoneChange,
	}
}

func (l *ChangeListener) Run(ctx context.Context) error {
	err := l.consumer.Consume(ctx, l.changeQueue, l.handleMessage)
	if err != nil {
		return err
	}

	<-ctx.Done()

	return ctx.Err()
}

func (l *ChangeListener) handleMessage(body []byte, headers messaging.Headers) error {
	l.logger.Info("received change")

	host, ok := headers.GetString(messaging.HeaderKeyHost)
	if !ok {
		return messaging.NewConsumerError(errors.New("host header not found"), false)
	}

	replyTo, ok := headers.GetString(messaging.HeaderKeyReplyTo)
	if !ok {
		return messaging.NewConsumerError(errors.New("reply-to header not found"), false)
	}

	var pbChange beacondnspb.Change
	if err := proto.Unmarshal(body, &pbChange); err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to unmarshal change: %w", err), false)
	}

	ch := convert.ChangeFromProto(&pbChange)

	var err error
	if ch.Type == model.ChangeTypeZone {
		err = l.onZoneChange(ch.ZoneChange)
	}

	if err != nil {
		return messaging.NewConsumerError(err, true)
	}

	err = l.AckChange(host, replyTo, &pbChange)
	if err != nil {
		return err
	}

	return nil
}

func (l *ChangeListener) AckChange(host string, replyTo string, change *beacondnspb.Change) error {
	headers := messaging.Headers{
		messaging.HeaderKeyHost: host,
	}

	ack := &beacondnspb.ChangeAck{
		ChangeId: change.GetId(),
	}

	body, err := proto.Marshal(ack)
	if err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to marshal change ack: %w", err), false)
	}

	err = l.publisher.Publish(context.Background(), replyTo, headers, body)
	if err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to publish change: %w", err), true)
	}

	return nil
}
