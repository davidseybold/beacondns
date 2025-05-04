package usecase

import (
	"context"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/google/uuid"
)

type ControllerService interface {
	CreateZone(ctx context.Context, name string) (*domain.CreateZoneResult, error)
	GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error)
	ListZones(ctx context.Context) ([]domain.Zone, error)

	ListResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]domain.ResourceRecordSet, error)
	ChangeResourceRecordSets(ctx context.Context, zoneID uuid.UUID, rrc domain.ChangeBatch) (*domain.ChangeInfo, error)

	AddNameServer(ctx context.Context, name string, routeKey string, ip string) (*domain.NameServer, error)
	ListNameServers(ctx context.Context) ([]domain.NameServer, error)
}
