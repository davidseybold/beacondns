package messaging

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpChannel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Close() error
	IsClosed() bool
	Qos(prefetchCount, prefetchSize int, global bool) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	QueueUnbind(name, key, exchange string, args amqp.Table) error
	QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error)
	QueuePurge(name string, noWait bool) (int, error)
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	ExchangeDelete(name string, ifUnused, noWait bool) error
	ExchangeBind(destination, key, source string, noWait bool, args amqp.Table) error
	ExchangeUnbind(destination, key, source string, noWait bool, args amqp.Table) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Cancel(consumer string, noWait bool) error
	Confirm(noWait bool) error
	NotifyPublish(confirm chan amqp.Confirmation) chan amqp.Confirmation
	NotifyReturn(ret chan amqp.Return) chan amqp.Return
	NotifyFlow(flow chan bool) chan bool
	NotifyClose(close chan *amqp.Error) chan *amqp.Error
	NotifyCancel(cancel chan string) chan string
	NotifyConfirm(ack, nack chan uint64) (chan uint64, chan uint64)
}

type amqpChannelAdapter struct {
	*amqp.Channel
}

var _ amqpChannel = (*amqpChannelAdapter)(nil)

func (a *amqpChannelAdapter) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return a.Channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (a *amqpChannelAdapter) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return a.Channel.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}

func (a *amqpChannelAdapter) Qos(prefetchCount, prefetchSize int, global bool) error {
	return a.Channel.Qos(prefetchCount, prefetchSize, global)
}

func (a *amqpChannelAdapter) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return a.Channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (a *amqpChannelAdapter) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return a.Channel.QueueBind(name, key, exchange, noWait, args)
}

func (a *amqpChannelAdapter) QueueUnbind(name, key, exchange string, args amqp.Table) error {
	return a.Channel.QueueUnbind(name, key, exchange, args)
}

func (a *amqpChannelAdapter) QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error) {
	return a.Channel.QueueDelete(name, ifUnused, ifEmpty, noWait)
}

func (a *amqpChannelAdapter) QueuePurge(name string, noWait bool) (int, error) {
	return a.Channel.QueuePurge(name, noWait)
}

func (a *amqpChannelAdapter) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return a.Channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

func (a *amqpChannelAdapter) ExchangeDelete(name string, ifUnused, noWait bool) error {
	return a.Channel.ExchangeDelete(name, ifUnused, noWait)
}

func (a *amqpChannelAdapter) ExchangeBind(destination, key, source string, noWait bool, args amqp.Table) error {
	return a.Channel.ExchangeBind(destination, key, source, noWait, args)
}

func (a *amqpChannelAdapter) ExchangeUnbind(destination, key, source string, noWait bool, args amqp.Table) error {
	return a.Channel.ExchangeUnbind(destination, key, source, noWait, args)
}

func (a *amqpChannelAdapter) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return a.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

func (a *amqpChannelAdapter) Cancel(consumer string, noWait bool) error {
	return a.Channel.Cancel(consumer, noWait)
}

func (a *amqpChannelAdapter) Confirm(noWait bool) error {
	return a.Channel.Confirm(noWait)
}

func (a *amqpChannelAdapter) NotifyPublish(confirm chan amqp.Confirmation) chan amqp.Confirmation {
	return a.Channel.NotifyPublish(confirm)
}

func (a *amqpChannelAdapter) NotifyReturn(ret chan amqp.Return) chan amqp.Return {
	return a.Channel.NotifyReturn(ret)
}

func (a *amqpChannelAdapter) NotifyFlow(flow chan bool) chan bool {
	return a.Channel.NotifyFlow(flow)
}

func (a *amqpChannelAdapter) NotifyClose(close chan *amqp.Error) chan *amqp.Error {
	return a.Channel.NotifyClose(close)
}

func (a *amqpChannelAdapter) NotifyCancel(cancel chan string) chan string {
	return a.Channel.NotifyCancel(cancel)
}

func (a *amqpChannelAdapter) NotifyConfirm(ack, nack chan uint64) (chan uint64, chan uint64) {
	return a.Channel.NotifyConfirm(ack, nack)
}

type amqpConnection interface {
	Channel() (amqpChannel, error)
	Close() error
}

type amqpConnectionAdapter struct {
	*amqp.Connection
}

func (a *amqpConnectionAdapter) Channel() (amqpChannel, error) {
	ch, err := a.Connection.Channel()
	if err != nil {
		return nil, err
	}
	return &amqpChannelAdapter{Channel: ch}, nil
}

type RabbitMQPublisher struct {
	channel      amqpChannel
	exchangeName string
	confirms     chan amqp.Confirmation
	close        chan *amqp.Error
}

var _ Publisher = (*RabbitMQPublisher)(nil)

func NewRabbitMQPublisher(conn *amqp.Connection, exchangeName string) (*RabbitMQPublisher, error) {
	adapter := &amqpConnectionAdapter{Connection: conn}
	channel, err := adapter.Channel()
	if err != nil {
		return nil, err
	}

	// Enable publisher confirms
	err = channel.Confirm(false)
	if err != nil {
		return nil, err
	}

	return &RabbitMQPublisher{
		channel:      channel,
		exchangeName: exchangeName,
		confirms:     channel.NotifyPublish(make(chan amqp.Confirmation, 1)),
		close:        channel.NotifyClose(make(chan *amqp.Error, 1)),
	}, nil
}

func (r *RabbitMQPublisher) Publish(ctx context.Context, routeKey string, headers Headers, message []byte) error {
	err := r.channel.Publish(
		r.exchangeName,
		routeKey,
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table(headers),
			Body:    message,
		},
	)
	if err != nil {
		return err
	}

	// Wait for confirmation
	select {
	case confirm := <-r.confirms:
		if !confirm.Ack {
			return fmt.Errorf("failed to publish message: %v", confirm)
		}
	case err := <-r.close:
		return fmt.Errorf("channel closed: %v", err)
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (r *RabbitMQPublisher) Close() error {
	return r.channel.Close()
}

type RabbitMQConsumerConfig struct {
	ConsumerName string
}

type RabbitMQConsumer struct {
	conn         amqpConnection
	autoAck      bool
	exclusive    bool
	noLocal      bool
	noWait       bool
	consumerName string
}

type RabbitMQConsumerHandler func(body []byte, headers Headers) error

func NewRabbitMQConsumer(consumerName string, conn *amqp.Connection) *RabbitMQConsumer {
	adapter := &amqpConnectionAdapter{Connection: conn}

	return &RabbitMQConsumer{
		conn:         adapter,
		consumerName: consumerName,
		autoAck:      false,
		exclusive:    true,
		noLocal:      true,
		noWait:       false,
	}
}

func (r *RabbitMQConsumer) Consume(ctx context.Context, queue string, handler RabbitMQConsumerHandler) error {
	channel, err := r.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	deliveries, err := channel.Consume(
		queue,
		r.consumerName,
		r.autoAck,
		r.exclusive,
		r.noLocal,
		r.noWait,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}
			err := handler(delivery.Body, Headers(delivery.Headers))
			if err != nil {
				if !r.autoAck {
					var consumerErr *ConsumerError
					if errors.As(err, &consumerErr) {
						delivery.Nack(false, consumerErr.IsRetryable())
					} else {
						// Default to retryable for unknown errors
						delivery.Nack(false, true)
					}
				}
				continue
			}

			if !r.autoAck {
				delivery.Ack(false)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type RabbitMQExchange struct {
	Name string
	Kind string
}

// RabbitMQTopology represents the exchange and queue configuration
type RabbitMQTopology struct {
	Exchange RabbitMQExchange
	Queues   []string
}

func SetupRabbitMQTopology(conn *amqp.Connection, topology RabbitMQTopology) error {
	adapter := &amqpConnectionAdapter{Connection: conn}
	channel, err := adapter.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		topology.Exchange.Name,
		topology.Exchange.Kind,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare and bind queues
	for _, queueName := range topology.Queues {
		_, err = channel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}

		err = channel.QueueBind(
			queueName,
			queueName,
			topology.Exchange.Name,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", queueName, err)
		}
	}

	return nil
}
