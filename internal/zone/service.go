package zone

import (
	"context"
	"errors"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

const (
	delegationSetSize = 2
	hostmasterEmail   = "hostmaster.beacondns.org."

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

	nameServerNames := []string{"ns1.beacondns.org.", "ns2.beacondns.org."}

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

	soaChangeAction := model.NewChangeAction(model.ChangeActionTypeUpsert, &soaRecord)
	nsChangeAction := model.NewChangeAction(model.ChangeActionTypeUpsert, &nsRecord)

	change := model.NewChange(zone.ID, model.ChangeStatusPending, []model.ChangeAction{soaChangeAction, nsChangeAction})

	event := NewCreateZoneEvent(zoneName, change.ID)

	var zoneInfo *model.ZoneInfo
	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var createZoneErr error
		zoneInfo, createZoneErr = r.GetZoneRepository().CreateZone(ctx, zone)
		if createZoneErr != nil {
			return createZoneErr
		}

		_, createZoneErr = r.GetZoneRepository().CreateChange(ctx, change)
		if createZoneErr != nil {
			return createZoneErr
		}

		createZoneErr = r.GetEventRepository().CreateEvent(ctx, event)
		if createZoneErr != nil {
			return createZoneErr
		}

		return nil
	})
	if err != nil && errors.Is(err, repository.ErrEntityAlreadyExists) {
		return nil, beaconerr.ErrZoneAlreadyExists("zone already exists")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to create zone", err)
	}

	return zoneInfo, nil
}

func (d *DefaultService) GetZoneInfo(ctx context.Context, name string) (*model.ZoneInfo, error) {
	zoneName := dns.Fqdn(name)
	z, err := d.registry.GetZoneRepository().GetZoneInfo(ctx, zoneName)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return nil, beaconerr.ErrNoSuchZone("zone not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get zone info", err)
	}

	return z, nil
}

func (d *DefaultService) ListZones(ctx context.Context) ([]model.ZoneInfo, error) {
	zones, err := d.registry.GetZoneRepository().ListZoneInfos(ctx)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to list zones", err)
	}

	return zones, nil
}

func (d *DefaultService) UpsertResourceRecordSet(
	ctx context.Context,
	zoneName string,
	rrSet *model.ResourceRecordSet,
) (*model.ResourceRecordSet, error) {
	zoneName = dns.Fqdn(zoneName)
	zone, err := d.registry.GetZoneRepository().GetZone(ctx, zoneName)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return nil, beaconerr.ErrNoSuchZone("zone not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to upsert resource record set", err)
	}

	rrSet.Name = dns.Fqdn(rrSet.Name)

	changeAction := model.NewChangeAction(model.ChangeActionTypeUpsert, rrSet)

	change := model.NewChange(zone.ID, model.ChangeStatusPending, []model.ChangeAction{changeAction})

	changeEvent := NewChangeRRSetEvent(change.ID.String(), change.ID)

	err = validateChanges(zone, &change)
	if err != nil {
		return nil, beaconerr.ErrInvalidArgument(err.Error(), "")
	}

	var newRRSet *model.ResourceRecordSet
	err = d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		newRRSet, err = r.GetZoneRepository().UpsertResourceRecordSet(ctx, zoneName, rrSet)
		if err != nil {
			return err
		}

		_, err = r.GetZoneRepository().CreateChange(ctx, change)
		if err != nil {
			return err
		}

		err = r.GetEventRepository().CreateEvent(ctx, changeEvent)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to upsert resource record set", err)
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
	rrSet, err := d.registry.GetZoneRepository().GetResourceRecordSet(ctx, zoneName, name, rrType)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return nil, beaconerr.ErrNoSuchResourceRecordSet("resource record set not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get resource record set", err)
	}
	return rrSet, nil
}

func (d *DefaultService) DeleteResourceRecordSet(
	ctx context.Context,
	zoneName string,
	name string,
	rrType model.RRType,
) error {
	zoneName = dns.Fqdn(zoneName)
	zone, err := d.registry.GetZoneRepository().GetZone(ctx, zoneName)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchZone("zone not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to delete resource record set", err)
	}

	changeAction := model.NewChangeAction(model.ChangeActionTypeDelete, nil)
	change := model.NewChange(zone.ID, model.ChangeStatusPending, []model.ChangeAction{changeAction})

	err = validateChanges(zone, &change)
	if err != nil {
		return beaconerr.ErrInvalidArgument(err.Error(), "")
	}

	changeEvent := NewChangeRRSetEvent(change.ID.String(), change.ID)

	err = d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		deleteErr := r.GetZoneRepository().DeleteResourceRecordSet(ctx, zoneName, name, rrType)
		if deleteErr != nil {
			return err
		}

		_, deleteErr = r.GetZoneRepository().CreateChange(ctx, change)
		if deleteErr != nil {
			return err
		}

		deleteErr = r.GetEventRepository().CreateEvent(ctx, changeEvent)
		if deleteErr != nil {
			return err
		}

		return nil
	})
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchResourceRecordSet("resource record set not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to delete zone", err)
	}

	return nil
}

func (d *DefaultService) ListResourceRecordSets(
	ctx context.Context,
	zoneName string,
) ([]model.ResourceRecordSet, error) {
	zoneName = dns.Fqdn(zoneName)
	rrSets, err := d.registry.GetZoneRepository().GetZoneResourceRecordSets(ctx, zoneName)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to list resource record sets", err)
	}

	return rrSets, nil
}

func (d *DefaultService) DeleteZone(ctx context.Context, name string) error {
	zoneName := dns.Fqdn(name)

	zone, err := d.registry.GetZoneRepository().GetZoneInfo(ctx, zoneName)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchZone("zone not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to delete zone", err)
	}

	if zone.ResourceRecordSetCount > 2 {
		return beaconerr.ErrHostedZoneNotEmpty("zone is not empty")
	}

	event := NewDeleteZoneEvent(zoneName)

	err = d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		deleteErr := r.GetZoneRepository().DeleteZone(ctx, zoneName)
		if deleteErr != nil {
			return deleteErr
		}

		deleteErr = r.GetEventRepository().CreateEvent(ctx, event)
		if deleteErr != nil {
			return deleteErr
		}

		return nil
	})
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchZone("zone not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to delete zone", err)
	}

	return nil
}
