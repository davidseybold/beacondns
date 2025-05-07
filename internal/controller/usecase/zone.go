package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/miekg/dns"
	"google.golang.org/protobuf/proto"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/controller/repository"
)

const (
	delegationSetSize = 2
	hostmasterEmail   = "hostmaster.beacondns.org"
)

type ZoneService interface {
	// Zone management
	CreateZone(ctx context.Context, name string) (*domain.CreateZoneResult, error)
	GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error)
	ListZones(ctx context.Context) ([]domain.Zone, error)

	// Resource record management
	ListResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]domain.ResourceRecordSet, error)
	ChangeResourceRecordSets(ctx context.Context, zoneID uuid.UUID, rrc domain.ChangeBatch) (*domain.ChangeInfo, error)
}

type DefaultZoneService struct {
	registry repository.TransactorRegistry
}

var _ ZoneService = (*DefaultZoneService)(nil)

func NewZoneService(r repository.TransactorRegistry) *DefaultZoneService {
	return &DefaultZoneService{
		registry: r,
	}
}

func (d *DefaultZoneService) CreateZone(ctx context.Context, name string) (*domain.CreateZoneResult, error) {
	zoneName := dns.Fqdn(name)

	zone := domain.Zone{
		ID:        uuid.New(),
		Name:      zoneName,
		IsPrivate: true,
	}

	var (
		ds              *domain.DelegationSet
		nameServerNames []string
	)

	if !zone.IsPrivate {
		nameServers, err := d.registry.GetNameServerRepository().GetRandomNameServers(ctx, delegationSetSize)
		if err != nil {
			return nil, err
		} else if len(nameServers) < delegationSetSize {
			return nil, errors.New("not enough name servers available")
		}

		ds = &domain.DelegationSet{
			ID:          uuid.New(),
			NameServers: nameServers,
		}

		for _, ns := range nameServers {
			nameServerNames = append(nameServerNames, ns.Name)
		}
	} else {
		ds = nil
		nameServerNames = []string{"ns00.beacondns.org.", "ns01.beacondns.org."}
	}

	primaryNS := nameServerNames[0]

	soa := domain.ResourceRecordSet{
		ID:   uuid.New(),
		Name: zoneName,
		Type: domain.RRTypeSOA,
		TTL:  900,
		ResourceRecords: []domain.ResourceRecord{
			{
				Value: fmt.Sprintf("%s %s %d %d %d %d %d", primaryNS, hostmasterEmail, 1, 7200, 900, 1209600, 86400),
			},
		},
	}

	nsRecRecords := make([]domain.ResourceRecord, len(nameServerNames))
	for i, name := range nameServerNames {
		nsRecRecords[i] = domain.ResourceRecord{
			Value: name,
		}
	}

	nsRec := domain.ResourceRecordSet{
		ID:              uuid.New(),
		Name:            zoneName,
		Type:            domain.RRTypeNS,
		TTL:             172800,
		ResourceRecords: nsRecRecords,
	}

	rrSetChanges := []domain.ResourceRecordSetChange{
		{
			Action:            domain.RRSetChangeActionCreate,
			ResourceRecordSet: soa,
		},
		{
			Action:            domain.RRSetChangeActionCreate,
			ResourceRecordSet: nsRec,
		},
	}

	change := domain.ZoneChange{
		ID:      uuid.New(),
		ZoneID:  zone.ID,
		Action:  domain.ZoneChangeActionCreateZone,
		Changes: rrSetChanges,
	}

	zoneChangeEvent := newZoneChangeEvent(zoneName, domain.ZoneChangeActionCreateZone, rrSetChanges)
	payload, err := proto.Marshal(zoneChangeEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal zone change event: %w", err)
	}

	msgs := make([]domain.OutboxMessage, len(ds.NameServers))
	for i, ns := range ds.NameServers {
		msgs[i] = domain.OutboxMessage{
			ID:       uuid.New(),
			Payload:  payload,
			RouteKey: fmt.Sprintf("nameserver.%s", ns.RouteKey),
		}
	}

	var syncs []domain.ZoneChangeSync
	for _, ns := range ds.NameServers {
		syncs = append(syncs, domain.ZoneChangeSync{
			ZoneChangeID: change.ID,
			NameServerID: ns.ID,
			Status:       domain.ChangeSyncStatusPending,
		})
	}

	params := repository.CreateZoneParams{
		Zone:          zone,
		DelegationSet: ds,
		SOA:           soa,
		NS:            nsRec,
		Change:        change,
		Syncs:         syncs,
	}

	createZoneFunc := func(ctx context.Context, r repository.Registry) (any, error) {
		out, err := r.GetZoneRepository().CreateZone(ctx, params)
		if err != nil {
			return nil, err
		}

		if err = r.GetOutboxRepository().InsertMessages(ctx, msgs); err != nil {
			return nil, err
		}

		return out, nil
	}

	rawResult, err := d.registry.WithinTransaction(ctx, createZoneFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create zone: %w", err)
	}

	changeInfo, ok := rawResult.(*domain.ChangeInfo)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	return &domain.CreateZoneResult{
		ZoneInfo: domain.ZoneInfo{
			Zone:          zone,
			DelegationSet: ds,
		},
		ChangeInfo: *changeInfo,
	}, nil
}

func (d *DefaultZoneService) GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error) {
	panic("unimplemented")
}

func (d *DefaultZoneService) ListZones(ctx context.Context) ([]domain.Zone, error) {
	panic("unimplemented")
}

func (d *DefaultZoneService) ChangeResourceRecordSets(ctx context.Context, zoneID uuid.UUID, b domain.ChangeBatch) (*domain.ChangeInfo, error) {
	panic("unimplemented")
}

func (d *DefaultZoneService) ListResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]domain.ResourceRecordSet, error) {
	panic("unimplemented")
}
