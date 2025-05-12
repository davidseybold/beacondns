package zone

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/miekg/dns"

	controllerdomain "github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/controller/repository"
	beacondomain "github.com/davidseybold/beacondns/internal/domain"
)

const (
	delegationSetSize = 2
	hostmasterEmail   = "hostmaster.beacondns.org"

	// SOA record values
	soaSerial     = 1
	soaRefresh    = 7200    // 2 hours
	soaRetry      = 900     // 15 minutes
	soaExpire     = 1209600 // 2 weeks
	soaMinimumTTL = 86400   // 1 day
)

type Service interface {
	// Zone management
	CreateZone(ctx context.Context, name string) (*controllerdomain.CreateZoneResult, error)
	GetZone(ctx context.Context, id uuid.UUID) (*controllerdomain.Zone, error)
	ListZones(ctx context.Context) ([]controllerdomain.Zone, error)

	// Resource record management
	ListResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]beacondomain.ResourceRecordSet, error)
	ChangeResourceRecordSets(ctx context.Context, zoneID uuid.UUID, rrc []beacondomain.ResourceRecordSetChange) (*controllerdomain.ChangeInfo, error)
}

type DefaultService struct {
	registry repository.TransactorRegistry
}

var _ Service = (*DefaultService)(nil)

func NewService(r repository.TransactorRegistry) *DefaultService {
	return &DefaultService{
		registry: r,
	}
}

func (d *DefaultService) CreateZone(ctx context.Context, name string) (*controllerdomain.CreateZoneResult, error) {
	zoneName := dns.Fqdn(name)

	zone := controllerdomain.Zone{
		ID:   uuid.New(),
		Name: zoneName,
	}

	nameServerNames := []string{"ns00.beacondns.org.", "ns01.beacondns.org."}

	primaryNS := nameServerNames[0]

	soa := beacondomain.NewSOA(zoneName, 900, primaryNS, hostmasterEmail, soaSerial, soaRefresh, soaRetry, soaExpire, soaMinimumTTL)

	nsRec := beacondomain.NewNS(zoneName, 172800, nameServerNames)

	params := repository.CreateZoneParams{
		Zone: zone,
		SOA:  soa,
		NS:   nsRec,
	}

	zoneChange := beacondomain.NewZoneChange(zoneName, beacondomain.ZoneChangeActionCreate, []beacondomain.ResourceRecordSetChange{
		beacondomain.NewResourceRecordSetChange(beacondomain.RRSetChangeActionCreate, soa),
		beacondomain.NewResourceRecordSetChange(beacondomain.RRSetChangeActionCreate, nsRec),
	})

	change := beacondomain.NewChangeWithZoneChange(zoneChange)

	targetServers, err := d.registry.GetServerRepository().GetAllServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get target servers: %w", err)
	}

	changeTargets := make([]controllerdomain.ChangeTarget, len(targetServers))

	for i, server := range targetServers {
		changeTargets[i] = controllerdomain.ChangeTarget{
			Server: *server,
			Status: controllerdomain.ChangeTargetStatusPending,
		}
	}

	changeWithTargets := controllerdomain.NewChangeWithTargets(change, changeTargets)

	createZoneFunc := func(ctx context.Context, r repository.Registry) (any, error) {
		err := r.GetZoneRepository().CreateZone(ctx, params)
		if err != nil {
			return nil, err
		}

		change, err := r.GetChangeRepository().CreateChange(ctx, changeWithTargets)
		if err != nil {
			return nil, err
		}

		return change, nil
	}

	rawResult, err := d.registry.WithinTransaction(ctx, createZoneFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create zone: %w", err)
	}

	chResult, ok := rawResult.(*controllerdomain.ChangeWithTargets)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	return &controllerdomain.CreateZoneResult{
		Zone: zone,
		ChangeInfo: controllerdomain.ChangeInfo{
			ID:          chResult.ID,
			Status:      controllerdomain.ChangeStatusPending,
			SubmittedAt: *chResult.SubmittedAt,
		},
	}, nil
}

func (d *DefaultService) GetZone(ctx context.Context, id uuid.UUID) (*controllerdomain.Zone, error) {
	panic("unimplemented")
}

func (d *DefaultService) ListZones(ctx context.Context) ([]controllerdomain.Zone, error) {
	panic("unimplemented")
}

func (d *DefaultService) ChangeResourceRecordSets(ctx context.Context, zoneID uuid.UUID, b []beacondomain.ResourceRecordSetChange) (*controllerdomain.ChangeInfo, error) {
	panic("unimplemented")
}

func (d *DefaultService) ListResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]beacondomain.ResourceRecordSet, error) {
	panic("unimplemented")
}
