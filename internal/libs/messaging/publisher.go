package messaging

import (
	"context"
)

type Publisher interface {
	Publish(ctx context.Context, exchange string, routeKey string, message []byte) error
	Close() error
}
