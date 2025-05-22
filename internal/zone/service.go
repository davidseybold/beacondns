package zone

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	CreateZone(ctx context.Context, name string) (*CreateZoneResult, error)
	DeleteZone(ctx context.Context, id uuid.UUID) (*model.Change, error)
	GetZoneInfo(ctx context.Context, id uuid.UUID) (*model.ZoneInfo, error)
	ListZones(ctx context.Context) ([]model.ZoneInfo, error)

	// Resource record management
	ListResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]model.ResourceRecordSet, error)
	ChangeResourceRecordSets(
		ctx context.Context,
		zoneID uuid.UUID,
		rrc []model.ResourceRecordSetChange,
	) (*model.Change, error)

	// Change management
	GetChange(ctx context.Context, id uuid.UUID) (*model.Change, error)
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

type CreateZoneResult struct {
	Zone   *model.Zone
	Change *model.Change
}

func (d *DefaultService) CreateZone(ctx context.Context, name string) (*CreateZoneResult, error) {
	var err error

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
			model.NewResourceRecordSetChange(model.RRSetChangeActionCreate, soaRecord),
			model.NewResourceRecordSetChange(model.RRSetChangeActionCreate, nsRecord),
		},
	)

	change := model.NewChangeWithZoneChange(zoneChange, model.ChangeStatusPending)

	rawResult, err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) (any, error) {
		var createZoneErr error
		createZoneErr = r.GetZoneRepository().CreateZone(ctx, &zone)
		if createZoneErr != nil {
			return nil, createZoneErr
		}

		change, createZoneErr := r.GetChangeRepository().CreateChange(ctx, change)
		if createZoneErr != nil {
			return nil, createZoneErr
		}

		return change, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create zone: %w", err)
	}

	chResult, ok := rawResult.(*model.Change)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	return &CreateZoneResult{
		Zone:   &zone,
		Change: chResult,
	}, nil
}

func (d *DefaultService) GetZoneInfo(ctx context.Context, id uuid.UUID) (*model.ZoneInfo, error) {
	z, err := d.registry.GetZoneRepository().GetZoneInfo(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, model.ErrZoneNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get zone %s: %w", id, err)
	}

	return z, nil
}

func (d *DefaultService) ListZones(ctx context.Context) ([]model.ZoneInfo, error) {
	return d.registry.GetZoneRepository().ListZoneInfos(ctx)
}

func (d *DefaultService) ChangeResourceRecordSets(
	ctx context.Context,
	zoneID uuid.UUID,
	rrc []model.ResourceRecordSetChange,
) (*model.Change, error) {
	zone, err := d.registry.GetZoneRepository().GetZone(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone %s: %w", zoneID, err)
	}

	err = validateChanges(zone, rrc)
	if err != nil {
		return nil, fmt.Errorf("failed to validate changes: %w", err)
	}

	zoneChange := model.NewZoneChange(
		zone.Name,
		model.ZoneChangeActionUpdate,
		rrc,
	)

	change := model.NewChangeWithZoneChange(
		zoneChange,
		model.ChangeStatusPending,
	)

	rawResult, err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) (any, error) {
		for _, ch := range rrc {
			switch ch.Action {
			case model.RRSetChangeActionCreate:
				ch.ResourceRecordSet.ID = uuid.New()
				err = r.GetZoneRepository().InsertResourceRecordSet(ctx, zoneID, &ch.ResourceRecordSet)
			case model.RRSetChangeActionUpsert:
				ch.ResourceRecordSet.ID = uuid.New()
				err = r.GetZoneRepository().UpsertResourceRecordSet(ctx, zoneID, &ch.ResourceRecordSet)
			case model.RRSetChangeActionDelete:
				err = r.GetZoneRepository().DeleteResourceRecordSet(ctx, zoneID, &ch.ResourceRecordSet)
			}

			if err != nil {
				return nil, err
			}
		}

		var newChange *model.Change
		newChange, err = r.GetChangeRepository().CreateChange(ctx, change)
		if err != nil {
			return nil, err
		}

		return newChange, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to change resource record sets: %w", err)
	}

	chResult, ok := rawResult.(*model.Change)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	return chResult, nil
}

func (d *DefaultService) ListResourceRecordSets(
	ctx context.Context,
	zoneID uuid.UUID,
) ([]model.ResourceRecordSet, error) {
	return d.registry.GetZoneRepository().GetZoneResourceRecordSets(ctx, zoneID)
}

func (d *DefaultService) DeleteZone(ctx context.Context, id uuid.UUID) (*model.Change, error) {
	zone, err := d.registry.GetZoneRepository().GetZone(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone %s: %w", id, err)
	}

	zoneChange := model.NewZoneChange(zone.Name, model.ZoneChangeActionDelete, []model.ResourceRecordSetChange{})
	change := model.NewChangeWithZoneChange(zoneChange, model.ChangeStatusPending)

	rawResult, err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) (any, error) {
		err := r.GetZoneRepository().DeleteZone(ctx, id)
		if err != nil {
			return nil, err
		}

		newChange, err := r.GetChangeRepository().CreateChange(ctx, change)
		if err != nil {
			return nil, err
		}

		return newChange, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete zone: %w", err)
	}

	chResult, ok := rawResult.(*model.Change)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rawResult)
	}

	return chResult, nil
}

func (d *DefaultService) GetChange(ctx context.Context, id uuid.UUID) (*model.Change, error) {
	ch, err := d.registry.GetChangeRepository().GetChange(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get change %s: %w", id, err)
	}

	return ch, nil
}
