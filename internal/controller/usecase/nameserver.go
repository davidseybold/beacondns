package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/controller/repository"
)

// NameServerService handles name server management operations
type NameServerService interface {
	AddNameServer(ctx context.Context, name string, routeKey string, ip string) (*domain.NameServer, error)
	ListNameServers(ctx context.Context) ([]domain.NameServer, error)
}

// DefaultNameServerService implements NameServerService
type DefaultNameServerService struct {
	registry repository.TransactorRegistry
}

var _ NameServerService = (*DefaultNameServerService)(nil)

// NewNameServerService creates a new instance of DefaultNameServerService
func NewNameServerService(r repository.TransactorRegistry) *DefaultNameServerService {
	return &DefaultNameServerService{
		registry: r,
	}
}

func (d *DefaultNameServerService) AddNameServer(ctx context.Context, name string, routeKey string, ip string) (*domain.NameServer, error) {
	ns := &domain.NameServer{
		ID:        uuid.New(),
		Name:      name,
		RouteKey:  routeKey,
		IPAddress: ip,
	}

	return d.registry.GetNameServerRepository().AddNameServer(ctx, ns)
}

func (d *DefaultNameServerService) ListNameServers(ctx context.Context) ([]domain.NameServer, error) {
	return d.registry.GetNameServerRepository().ListNameServers(ctx)
}
