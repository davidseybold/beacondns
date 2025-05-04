package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/controller/repository"
	"github.com/google/uuid"
	"github.com/miekg/dns"
)

const (
	delegationSetSize = 2
	hostmasterEmail   = "hostmaster.beacondns.org"
)

type DefaultControllerService struct {
	repo repository.BeaconDBRepository
}

var _ ControllerService = (*DefaultControllerService)(nil)

func NewControllerService(repo repository.BeaconDBRepository) *DefaultControllerService {
	return &DefaultControllerService{
		repo: repo,
	}
}

func (d *DefaultControllerService) AddNameServer(ctx context.Context, name string, routeKey string, ip string) (*domain.NameServer, error) {
	ns := &domain.NameServer{
		ID:        uuid.New(),
		Name:      name,
		RouteKey:  routeKey,
		IPAddress: ip,
	}

	return d.repo.AddNameServer(ctx, ns)
}

func (d *DefaultControllerService) ListNameServers(ctx context.Context) ([]domain.NameServer, error) {
	return d.repo.ListNameServers(ctx)
}

func (d *DefaultControllerService) CreateZone(ctx context.Context, name string) (*domain.CreateZoneResult, error) {

	zoneName := dns.Fqdn(name)

	zone := domain.Zone{
		ID:   uuid.New(),
		Name: zoneName,
	}

	nameServers, err := d.repo.GetRandomNameServers(ctx, delegationSetSize)
	if err != nil {
		return nil, err
	} else if len(nameServers) < delegationSetSize {
		return nil, errors.New("not enough name servers available")
	}

	ds := domain.DelegationSet{
		ID:          uuid.New(),
		NameServers: nameServers,
	}

	primaryNS := nameServers[0]

	soa := domain.ResourceRecordSet{
		ID:   uuid.New(),
		Name: zoneName,
		Type: domain.RRTypeSOA,
		TTL:  900,
		ResourceRecords: []domain.ResourceRecord{
			{
				Value: fmt.Sprintf("%s %s %d %d %d %d %d", primaryNS.Name, hostmasterEmail, 1, 7200, 900, 1209600, 86400),
			},
		},
	}

	nsRecRecords := make([]domain.ResourceRecord, len(nameServers))
	for i, ns := range nameServers {
		nsRecRecords[i] = domain.ResourceRecord{
			Value: ns.Name,
		}
	}

	nsRec := domain.ResourceRecordSet{
		ID:              uuid.New(),
		Name:            zoneName,
		Type:            domain.RRTypeNS,
		TTL:             172800,
		ResourceRecords: nsRecRecords,
	}

	change := domain.ZoneChange{
		ID:     uuid.New(),
		ZoneID: zone.ID,
		Action: domain.ZoneChangeActionCreateZone,
		Changes: []domain.ResourceRecordSetChange{
			{
				Action:            domain.RRSetChangeActionCreate,
				ResourceRecordSet: soa,
			},
			{
				Action:            domain.RRSetChangeActionCreate,
				ResourceRecordSet: nsRec,
			},
		},
	}

	payload := []byte(`{ "test": "payload" }`)

	msgs := make([]domain.OutboxMessage, len(ds.NameServers))
	for i, ns := range ds.NameServers {
		msgs[i] = domain.OutboxMessage{
			ID:       uuid.New(),
			Payload:  payload,
			RouteKey: ns.RouteKey,
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
		Zone:           &zone,
		DelegationSet:  &ds,
		SOA:            &soa,
		NS:             &nsRec,
		Change:         &change,
		OutboxMessages: msgs,
		Syncs:          syncs,
	}

	changeInfo, err := d.repo.CreateZone(ctx, params)
	if err != nil {
		return nil, err
	}

	return &domain.CreateZoneResult{
		ZoneInfo: domain.ZoneInfo{
			Zone:          zone,
			DelegationSet: ds,
		},
		ChangeInfo: *changeInfo,
	}, nil
}

func (d *DefaultControllerService) GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error) {
	panic("unimplemented")
}

func (d *DefaultControllerService) ListZones(ctx context.Context) ([]domain.Zone, error) {
	panic("unimplemented")
}

func (d *DefaultControllerService) ChangeResourceRecordSets(ctx context.Context, zoneID uuid.UUID, b domain.ChangeBatch) (*domain.ChangeInfo, error) {
	panic("unimplemented")
}

func (d *DefaultControllerService) ListResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]domain.ResourceRecordSet, error) {
	panic("unimplemented")
}
