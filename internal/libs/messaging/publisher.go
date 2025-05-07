package messaging

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, exchange string, routeKey string, message []byte) error
	Close() error
}

type RabbitMQPublisher struct {
	channel *amqp.Channel
}

func NewRabbitMQPublisher(conn *amqp.Connection) (*RabbitMQPublisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQPublisher{
		channel: channel,
	}, nil
}

func (r *RabbitMQPublisher) Publish(ctx context.Context, exchange string, routeKey string, message []byte) error {
	return r.channel.PublishWithContext(
		ctx,
		exchange, // exchange
		routeKey, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			Body: message,
		},
	)
}

func (r *RabbitMQPublisher) Close() error {
	return r.channel.Close()
}
