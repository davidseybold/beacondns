package zone

import (
	"context"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
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
	CreateZone(ctx context.Context, name string) (*model.ZoneInfo, error)
	DeleteZone(ctx context.Context, name string) error
	GetZoneInfo(ctx context.Context, name string) (*model.ZoneInfo, error)
	ListZones(ctx context.Context) ([]model.ZoneInfo, error)

	// Resource record management
	ListResourceRecordSets(ctx context.Context, zoneName string) ([]model.ResourceRecordSet, error)
	GetResourceRecordSet(
		ctx context.Context,
		zoneName string,
		name string,
		rrType model.RRType,
	) (*model.ResourceRecordSet, error)
	UpsertResourceRecordSet(
		ctx context.Context,
		zoneName string,
		rrSet *model.ResourceRecordSet,
	) (*model.ResourceRecordSet, error)
	DeleteResourceRecordSet(ctx context.Context, zoneName string, name string, rrType model.RRType) error
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

func (d *DefaultService) CreateZone(ctx context.Context, name string) (*model.ZoneInfo, error) {
	zoneName := dns.Fqdn(name)

	zone := model.NewZone(zoneName)

	nameServerNames := []string{"ns00.beacondns.org.", "ns01.beacondns.org."}

	primaryNS := nameServerNames[0]

	soaRecord := model.NewSOA(
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
	nsRecord := model.NewNS(zoneName, nsRRTTL, nameServerNames)

	zone.ResourceRecordSets = []model.ResourceRecordSet{soaRecord, nsRecord}

	zoneChange := model.NewZoneChange(
		zoneName,
		model.ZoneChangeActionCreate,
		[]model.ResourceRecordSetChange{
			model.NewResourceRecordSetChange(model.RRSetChangeActionUpsert, soaRecord),
			model.NewResourceRecordSetChange(model.RRSetChangeActionUpsert, nsRecord),
		},
	)

	change := model.NewChangeWithZoneChange(zoneChange, model.ChangeStatusPending)

	var zoneInfo *model.ZoneInfo
	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var createZoneErr error
		zoneInfo, createZoneErr = r.GetZoneRepository().CreateZone(ctx, zone)
		if createZoneErr != nil {
			return createZoneErr
		}

		_, createZoneErr = r.GetChangeRepository().CreateChange(ctx, change)
		if createZoneErr != nil {
			return createZoneErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return zoneInfo, nil
}

func (d *DefaultService) GetZoneInfo(ctx context.Context, name string) (*model.ZoneInfo, error) {
	zoneName := dns.Fqdn(name)
	z, err := d.registry.GetZoneRepository().GetZoneInfo(ctx, zoneName)
	if err != nil {
		return nil, err
	}

	return z, nil
}

func (d *DefaultService) ListZones(ctx context.Context) ([]model.ZoneInfo, error) {
	return d.registry.GetZoneRepository().ListZoneInfos(ctx)
}

func (d *DefaultService) UpsertResourceRecordSet(
	ctx context.Context,
	zoneName string,
	rrSet *model.ResourceRecordSet,
) (*model.ResourceRecordSet, error) {
	zoneName = dns.Fqdn(zoneName)
	zone, err := d.registry.GetZoneRepository().GetZone(ctx, zoneName)
	if err != nil {
		return nil, err
	}

	rrSet.Name = dns.Fqdn(rrSet.Name)

	rrcs := []model.ResourceRecordSetChange{
		model.NewResourceRecordSetChange(model.RRSetChangeActionUpsert, *rrSet),
	}

	zoneChange := model.NewZoneChange(
		zone.Name,
		model.ZoneChangeActionUpdate,
		rrcs,
	)

	change := model.NewChangeWithZoneChange(
		zoneChange,
		model.ChangeStatusPending,
	)

	err = validateChanges(zone, rrcs)
	if err != nil {
		return nil, err
	}

	var newRRSet *model.ResourceRecordSet
	err = d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		newRRSet, err = r.GetZoneRepository().UpsertResourceRecordSet(ctx, zoneName, rrSet)
		if err != nil {
			return err
		}

		_, err = r.GetChangeRepository().CreateChange(ctx, change)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newRRSet, nil
}

func (d *DefaultService) GetResourceRecordSet(
	ctx context.Context,
	zoneName string,
	name string,
	rrType model.RRType,
) (*model.ResourceRecordSet, error) {
	zoneName = dns.Fqdn(zoneName)
	return d.registry.GetZoneRepository().GetResourceRecordSet(ctx, zoneName, name, rrType)
}

func (d *DefaultService) DeleteResourceRecordSet(
	ctx context.Context,
	zoneName string,
	name string,
	rrType model.RRType,
) error {
	zoneName = dns.Fqdn(zoneName)
	return d.registry.GetZoneRepository().DeleteResourceRecordSet(ctx, zoneName, name, rrType)
}

func (d *DefaultService) ListResourceRecordSets(
	ctx context.Context,
	zoneName string,
) ([]model.ResourceRecordSet, error) {
	zoneName = dns.Fqdn(zoneName)
	return d.registry.GetZoneRepository().GetZoneResourceRecordSets(ctx, zoneName)
}

func (d *DefaultService) DeleteZone(ctx context.Context, name string) error {
	zoneName := dns.Fqdn(name)

	zoneChange := model.NewZoneChange(zoneName, model.ZoneChangeActionDelete, []model.ResourceRecordSetChange{})
	change := model.NewChangeWithZoneChange(zoneChange, model.ChangeStatusPending)

	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		err := r.GetZoneRepository().DeleteZone(ctx, zoneName)
		if err != nil {
			return err
		}

		_, err = r.GetChangeRepository().CreateChange(ctx, change)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
