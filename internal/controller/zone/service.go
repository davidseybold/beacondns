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

	soaSerial     = 1
	soaRefresh    = 7200    // 2 hours
	soaRetry      = 900     // 15 minutes
	soaExpire     = 1209600 // 2 weeks
	soaMinimumTTL = 86400   // 1 day

	soaRRTTL = 86400  // 1 day
	nsRRTTL  = 172800 // 48 hours
)

type Service interface {
	// Zone management
	CreateZone(ctx context.Context, name string) (*controllerdomain.CreateZoneResult, error)
	GetZone(ctx context.Context, id uuid.UUID) (*controllerdomain.Zone, error)
	ListZones(ctx context.Context) ([]controllerdomain.Zone, error)

	// Resource record management
	ListResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]beacondomain.ResourceRecordSet, error)
	ChangeResourceRecordSets(
		ctx context.Context,
		zoneID uuid.UUID,
		rrc []beacondomain.ResourceRecordSetChange,
	) (*controllerdomain.ChangeInfo, error)
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
	var err error

	zoneName := dns.Fqdn(name)

	zone := controllerdomain.Zone{
		ID:   uuid.New(),
		Name: zoneName,
	}

	nameServerNames := []string{"ns00.beacondns.org.", "ns01.beacondns.org."}

	primaryNS := nameServerNames[0]

	soa := beacondomain.NewSOA(
		zoneName,
		soaRRTTL,
		primaryNS,
		hostmasterEmail,
		soaSerial,
		soaRefresh,
		soaRetry,
		soaExpire,
		soaMinimumTTL,
	)
	nsRec := beacondomain.NewNS(zoneName, nsRRTTL, nameServerNames)

	params := repository.CreateZoneParams{
		Zone: zone,
		SOA:  soa,
		NS:   nsRec,
	}

	zoneChange := beacondomain.NewZoneChange(
		zoneName,
		beacondomain.ZoneChangeActionCreate,
		[]beacondomain.ResourceRecordSetChange{
			beacondomain.NewResourceRecordSetChange(beacondomain.RRSetChangeActionCreate, soa),
			beacondomain.NewResourceRecordSetChange(beacondomain.RRSetChangeActionCreate, nsRec),
		},
	)

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
		var createZoneErr error
		createZoneErr = r.GetZoneRepository().CreateZone(ctx, params)
		if createZoneErr != nil {
			return nil, createZoneErr
		}

		change, createZoneErr := r.GetChangeRepository().CreateChange(ctx, changeWithTargets)
		if createZoneErr != nil {
			return nil, createZoneErr
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

func (d *DefaultService) GetZone(_ context.Context, _ uuid.UUID) (*controllerdomain.Zone, error) {
	panic("unimplemented")
}

func (d *DefaultService) ListZones(_ context.Context) ([]controllerdomain.Zone, error) {
	panic("unimplemented")
}

func (d *DefaultService) ChangeResourceRecordSets(
	_ context.Context,
	_ uuid.UUID,
	_ []beacondomain.ResourceRecordSetChange,
) (*controllerdomain.ChangeInfo, error) {
	panic("unimplemented")
}

func (d *DefaultService) ListResourceRecordSets(
	_ context.Context,
	_ uuid.UUID,
) ([]beacondomain.ResourceRecordSet, error) {
	panic("unimplemented")
}
