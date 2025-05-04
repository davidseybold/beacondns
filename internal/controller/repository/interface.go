package repository

import (
	"context"

	"github.com/davidseybold/beacondns/internal/controller/domain"
)

type BeaconDBRepository interface {
	AddNameServer(ctx context.Context, ns *domain.NameServer) (*domain.NameServer, error)
	ListNameServers(ctx context.Context) ([]domain.NameServer, error)

	GetRandomNameServers(ctx context.Context, count int) ([]domain.NameServer, error)

	CreateZone(ctx context.Context, params CreateZoneParams) (*domain.ChangeInfo, error)
}

type CreateZoneParams struct {
	Zone           *domain.Zone
	DelegationSet  *domain.DelegationSet
	SOA            *domain.ResourceRecordSet
	NS             *domain.ResourceRecordSet
	Change         *domain.ZoneChange
	OutboxMessages []domain.OutboxMessage
	Syncs          []domain.ZoneChangeSync
}
