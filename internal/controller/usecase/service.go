package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/controller/repository"
	beacondnspb "github.com/davidseybold/beacondns/internal/libs/gen/proto/beacondns/v1"
)

const (
	delegationSetSize = 2
	hostmasterEmail   = "hostmaster.beacondns.org"
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

type DefaultControllerService struct {
	registry repository.TransactorRegistry
}

var _ ControllerService = (*DefaultControllerService)(nil)

func NewControllerService(r repository.TransactorRegistry) *DefaultControllerService {
	return &DefaultControllerService{
		registry: r,
	}
}

func (d *DefaultControllerService) AddNameServer(ctx context.Context, name string, routeKey string, ip string) (*domain.NameServer, error) {
	ns := &domain.NameServer{
		ID:        uuid.New(),
		Name:      name,
		RouteKey:  routeKey,
		IPAddress: ip,
	}

	return d.registry.GetNameServerRepository().AddNameServer(ctx, ns)
}

func (d *DefaultControllerService) ListNameServers(ctx context.Context) ([]domain.NameServer, error) {
	return d.registry.GetNameServerRepository().ListNameServers(ctx)
}

func (d *DefaultControllerService) CreateZone(ctx context.Context, name string) (*domain.CreateZoneResult, error) {

	zoneName := dns.Fqdn(name)

	zone := domain.Zone{
		ID:   uuid.New(),
		Name: zoneName,
	}

	nameServers, err := d.registry.GetNameServerRepository().GetRandomNameServers(ctx, delegationSetSize)
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

var domainZoneChangeActionToProto = map[domain.ZoneChangeAction]beacondnspb.ZoneChangeAction{
	domain.ZoneChangeActionCreateZone: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_CREATE_ZONE,
	domain.ZoneChangeActionUpdateZone: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UPDATE_ZONE,
	domain.ZoneChangeActionDeleteZone: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_DELETE_ZONE,
}

var domainRRSetChangeActionToProto = map[domain.RRSetChangeAction]beacondnspb.RRSetChangeAction{
	domain.RRSetChangeActionCreate: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE,
	domain.RRSetChangeActionUpsert: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UPSERT,
	domain.RRSetChangeActionDelete: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_DELETE,
}

var domainRRTypeToProto = map[domain.RRType]beacondnspb.RRType{
	domain.RRTypeSOA:    beacondnspb.RRType_RR_TYPE_SOA,
	domain.RRTypeNS:     beacondnspb.RRType_RR_TYPE_NS,
	domain.RRTypeA:      beacondnspb.RRType_RR_TYPE_A,
	domain.RRTypeAAAA:   beacondnspb.RRType_RR_TYPE_AAAA,
	domain.RRTypeCNAME:  beacondnspb.RRType_RR_TYPE_CNAME,
	domain.RRTypeMX:     beacondnspb.RRType_RR_TYPE_MX,
	domain.RRTypeTXT:    beacondnspb.RRType_RR_TYPE_TXT,
	domain.RRTypeSRV:    beacondnspb.RRType_RR_TYPE_SRV,
	domain.RRTypePTR:    beacondnspb.RRType_RR_TYPE_PTR,
	domain.RRTypeCAA:    beacondnspb.RRType_RR_TYPE_CAA,
	domain.RRTypeNAPTR:  beacondnspb.RRType_RR_TYPE_NAPTR,
	domain.RRTypeDS:     beacondnspb.RRType_RR_TYPE_DS,
	domain.RRTypeDNSKEY: beacondnspb.RRType_RR_TYPE_DNSKEY,
	domain.RRTypeRRSIG:  beacondnspb.RRType_RR_TYPE_RRSIG,
	domain.RRTypeNSEC:   beacondnspb.RRType_RR_TYPE_NSEC,
	domain.RRTypeTLSA:   beacondnspb.RRType_RR_TYPE_TLSA,
	domain.RRTypeSPF:    beacondnspb.RRType_RR_TYPE_SPF,
}

func newZoneChangeEvent(zoneName string, action domain.ZoneChangeAction, changes []domain.ResourceRecordSetChange) *beacondnspb.ZoneChangeEvent {
	pbZoneAction, ok := domainZoneChangeActionToProto[action]
	if !ok {
		pbZoneAction = beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UNSPECIFIED
	}

	pbChanges := make([]*beacondnspb.ResourceRecordSetChange, len(changes))
	for i, change := range changes {
		pbAction, ok := domainRRSetChangeActionToProto[change.Action]
		if !ok {
			pbAction = beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UNSPECIFIED
		}

		pbType, ok := domainRRTypeToProto[change.ResourceRecordSet.Type]
		if !ok {
			pbType = beacondnspb.RRType_RR_TYPE_UNSPECIFIED
		}

		records := make([]*beacondnspb.ResourceRecord, len(change.ResourceRecordSet.ResourceRecords))
		for j, rr := range change.ResourceRecordSet.ResourceRecords {
			records[j] = &beacondnspb.ResourceRecord{
				Value: rr.Value,
			}
		}

		pbChanges[i] = &beacondnspb.ResourceRecordSetChange{
			Action: pbAction,
			ResourceRecordSet: &beacondnspb.ResourceRecordSet{
				Name:            change.ResourceRecordSet.Name,
				Type:            pbType,
				Ttl:             uint32(change.ResourceRecordSet.TTL),
				ResourceRecords: records,
			},
		}
	}

	return &beacondnspb.ZoneChangeEvent{
		ZoneName: zoneName,
		Action:   pbZoneAction,
		Changes:  pbChanges,
	}
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
