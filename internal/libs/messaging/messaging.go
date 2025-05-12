package messaging

import (
	"context"
)

const HeaderKeyHost = "x-host"
const HeaderKeyReplyTo = "x-reply-to"

type Publisher interface {
	Publish(ctx context.Context, routeKey string, headers Headers, message []byte) error
	Close() error
}

type Consumer interface {
	Consume(ctx context.Context, queue string, handler RabbitMQConsumerHandler) error
}

type Headers map[string]any

func (h Headers) GetString(key string) (string, bool) {
	value, ok := h[key]
	if !ok {
		return "", false
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", false
	}

	return valueStr, true
}

type ConsumerError struct {
	err       error
	retryable bool
}

func NewConsumerError(err error, retryable bool) *ConsumerError {
	return &ConsumerError{
		err:       err,
		retryable: retryable,
	}
}

func (e *ConsumerError) Error() string {
	return e.err.Error()
}

func (e *ConsumerError) IsRetryable() bool {
	return e.retryable
}

func (e *ConsumerError) Unwrap() error {
	return e.err
}
