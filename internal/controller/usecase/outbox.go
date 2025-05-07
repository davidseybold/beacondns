package usecase

import (
	"context"

	"github.com/davidseybold/beacondns/internal/controller/repository"
	"github.com/davidseybold/beacondns/internal/libs/messaging"
	"github.com/google/uuid"
)

const (
	beaconExchange = "beacon"
)

type OutboxService interface {
	ProcessNextBatch(ctx context.Context, batchSize int) (int, error)
}

var _ OutboxService = (*DefaultOutboxService)(nil)

type DefaultOutboxService struct {
	repoRegistry repository.Registry
	publisher    messaging.Publisher
}

func NewOutboxService(reg repository.Registry, publisher messaging.Publisher) *DefaultOutboxService {
	return &DefaultOutboxService{
		repoRegistry: reg,
		publisher:    publisher,
	}
}

func (d *DefaultOutboxService) ProcessNextBatch(ctx context.Context, batchSize int) (int, error) {
	pendingMsgs, err := d.repoRegistry.GetOutboxRepository().GetPendingMessages(context.Background(), batchSize)
	if err != nil {
		return -1, err
	}

	if len(pendingMsgs) == 0 {
		return 0, nil
	}

	msgsIDsToDelete := make([]uuid.UUID, 0, len(pendingMsgs))

	for _, msg := range pendingMsgs {
		if err := d.publisher.Publish(ctx, beaconExchange, msg.RouteKey, msg.Payload); err != nil {
			continue
		}
		msgsIDsToDelete = append(msgsIDsToDelete, msg.ID)

	}

	if err := d.repoRegistry.GetOutboxRepository().DeleteMessages(context.Background(), msgsIDsToDelete); err != nil {
		return len(msgsIDsToDelete), err
	}

	return len(msgsIDsToDelete), nil
}
