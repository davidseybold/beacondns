package resolver

import (
	"context"
	"errors"
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"

	beacondnspb "github.com/davidseybold/beacondns/internal/libs/gen/proto/beacondns/v1"
	"github.com/davidseybold/beacondns/internal/libs/messaging"
)

type ChangeListenerConfig struct {
	Consumer    messaging.Consumer
	Publisher   messaging.Publisher
	ChangeQueue string
}

type ChangeListener struct {
	ctx         context.Context
	consumer    messaging.Consumer
	publisher   messaging.Publisher
	changeQueue string
}

func NewChangeListener(ctx context.Context, config ChangeListenerConfig) *ChangeListener {
	return &ChangeListener{
		ctx:         ctx,
		consumer:    config.Consumer,
		publisher:   config.Publisher,
		changeQueue: config.ChangeQueue,
	}
}

func (l *ChangeListener) Run() error {
	err := l.consumer.Consume(l.ctx, l.changeQueue, l.handleMessage)
	if err != nil {
		return err
	}

	<-l.ctx.Done()

	return l.ctx.Err()
}

func (l *ChangeListener) handleMessage(body []byte, headers messaging.Headers) error {

	log.Println("received change")

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

	err := l.AckChange(host, replyTo, &pbChange)
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
		ChangeId: change.Id,
	}

	body, err := proto.Marshal(ack)
	if err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to marshal change ack: %w", err), false)
	}

	err = l.publisher.Publish(l.ctx, replyTo, headers, body)
	if err != nil {
		return messaging.NewConsumerError(fmt.Errorf("failed to publish change: %w", err), true)
	}

	return nil
}
