package messaging

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPChannel interface {
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
	Consume(
		queue,
		consumer string,
		autoAck,
		exclusive,
		noLocal,
		noWait bool,
		args amqp.Table,
	) (<-chan amqp.Delivery, error)
	Cancel(consumer string, noWait bool) error
	Confirm(noWait bool) error
	NotifyPublish(confirm chan amqp.Confirmation) chan amqp.Confirmation
	NotifyReturn(ret chan amqp.Return) chan amqp.Return
	NotifyFlow(flow chan bool) chan bool
	NotifyClose(c chan *amqp.Error) chan *amqp.Error
	NotifyCancel(cancel chan string) chan string
	NotifyConfirm(ack, nack chan uint64) (chan uint64, chan uint64)
}

type AMQPChannelAdapter struct {
	*amqp.Channel
}

var _ AMQPChannel = (*AMQPChannelAdapter)(nil)

type AMQPConnection interface {
	Channel() (AMQPChannel, error)
	Close() error
}

type AMQPConnectionAdapter struct {
	*amqp.Connection
}

func (a *AMQPConnectionAdapter) Channel() (AMQPChannel, error) {
	ch, err := a.Connection.Channel()
	if err != nil {
		return nil, err
	}
	return &AMQPChannelAdapter{Channel: ch}, nil
}

func DialAMQP(connString string) (AMQPConnection, error) {
	conn, err := amqp.Dial(connString)
	if err != nil {
		return nil, err
	}
	return &AMQPConnectionAdapter{Connection: conn}, nil
}

type RabbitMQPublisher struct {
	conn         AMQPConnection
	channel      AMQPChannel
	exchangeName string
	confirms     chan amqp.Confirmation
	close        chan *amqp.Error
}

var _ Publisher = (*RabbitMQPublisher)(nil)

func NewRabbitMQPublisher(conn AMQPConnection, exchangeName string) (*RabbitMQPublisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Enable publisher confirms
	err = channel.Confirm(false)
	if err != nil {
		return nil, err
	}

	return &RabbitMQPublisher{
		conn:         conn,
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
		return fmt.Errorf("channel closed: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

type RabbitMQConsumerConfig struct {
	ConsumerName string
}

type RabbitMQConsumer struct {
	conn         AMQPConnection
	autoAck      bool
	exclusive    bool
	noLocal      bool
	noWait       bool
	consumerName string
}

type RabbitMQConsumerHandler func(body []byte, headers Headers) error

func NewRabbitMQConsumer(consumerName string, conn AMQPConnection) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		conn:         conn,
		consumerName: consumerName,
		autoAck:      false,
		exclusive:    true,
		noLocal:      true,
		noWait:       false,
	}
}

func (r *RabbitMQConsumer) Consume(ctx context.Context, queue string, handler RabbitMQConsumerHandler) error {
	var err error
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
			if err = handler(delivery.Body, Headers(delivery.Headers)); err != nil {
				r.handleDeliveryError(delivery, err)
				continue
			}
			if !r.autoAck {
				_ = delivery.Ack(false)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (r *RabbitMQConsumer) handleDeliveryError(delivery amqp.Delivery, err error) {
	if r.autoAck {
		return
	}

	var consumerErr *ConsumerError
	isRetryable := true
	if errors.As(err, &consumerErr) {
		isRetryable = consumerErr.IsRetryable()
	}
	_ = delivery.Nack(false, isRetryable)
}

type RabbitMQExchange struct {
	Name string
	Kind string
}

type RabbitMQTopology struct {
	Exchange RabbitMQExchange
	Queues   []string
}

func SetupAMQPTopology(conn AMQPConnection, topology RabbitMQTopology) error {
	channel, err := conn.Channel()
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
