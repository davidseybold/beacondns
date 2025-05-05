package usecase

import (
	"context"

	"github.com/davidseybold/beacondns/internal/controller/domain"
)

type OutboxService interface {
	SendOutboxMessage(ctx context.Context, msg domain.OutboxMessage) error
}

var _ OutboxService = (*DefaultOutboxService)(nil)

type DefaultOutboxService struct {
}

func NewOutboxService() *DefaultOutboxService {
	return &DefaultOutboxService{}
}

func (d *DefaultOutboxService) SendOutboxMessage(ctx context.Context, msg domain.OutboxMessage) error {
	// This method is a placeholder for sending outbox messages.
	// In a real implementation, this would involve writing the message to a message queue.
	// For now, we will just return nil to indicate success.
	return nil
}
