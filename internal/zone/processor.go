package zone

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/davidseybold/beacondns/internal/dns"
	"github.com/davidseybold/beacondns/internal/dnsstore"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

type EventProcessor struct {
	repository repository.TransactorRegistry
	store      dnsstore.ZoneWriter
	logger     *slog.Logger
}

type EventProcessorDeps struct {
	Repository repository.TransactorRegistry
	DNSStore   dnsstore.ZoneWriter
	Logger     *slog.Logger
}

func (d *EventProcessorDeps) Validate() error {
	if d.Repository == nil {
		return errors.New("repository is required")
	}

	if d.DNSStore == nil {
		return errors.New("dns store is required")
	}

	if d.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

func NewEventProcessor(deps *EventProcessorDeps) (*EventProcessor, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return &EventProcessor{
		repository: deps.Repository,
		store:      deps.DNSStore,
		logger:     deps.Logger,
	}, nil
}

func (p *EventProcessor) Events() []string {
	return []string{
		EventTypeCreateZone,
		EventTypeDeleteZone,
		EventTypeChangeRRSet,
	}
}

func (p *EventProcessor) ProcessEvent(ctx context.Context, event *model.Event) error {
	switch event.Type {
	case EventTypeCreateZone:
		return p.processCreateZoneEvent(ctx, event)
	case EventTypeDeleteZone:
		return p.processDeleteZoneEvent(ctx, event)
	case EventTypeChangeRRSet:
		return p.processChangeRRSetEvent(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (p *EventProcessor) processCreateZoneEvent(ctx context.Context, event *model.Event) error {
	var createZoneEvent CreateZoneEvent
	if err := json.Unmarshal(event.Payload, &createZoneEvent); err != nil {
		return err
	}

	change, err := p.repository.GetZoneRepository().GetChange(ctx, createZoneEvent.ChangeID)
	if err != nil {
		return err
	}

	tx := p.store.ZoneTxn(ctx, createZoneEvent.ZoneName)

	tx.CreateZoneMarker()

	for _, changeAction := range change.Actions {
		if err = processChangeAction(tx, changeAction); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = p.repository.GetZoneRepository().UpdateChangeStatus(ctx, change.ID, model.ChangeStatusDone)
	if err != nil {
		return err
	}

	return nil
}

func (p *EventProcessor) processDeleteZoneEvent(ctx context.Context, event *model.Event) error {
	var deleteZoneEvent DeleteZoneEvent
	if err := json.Unmarshal(event.Payload, &deleteZoneEvent); err != nil {
		return err
	}

	return p.store.DeleteZone(ctx, deleteZoneEvent.ZoneName)
}

func (p *EventProcessor) processChangeRRSetEvent(ctx context.Context, event *model.Event) error {
	var changeRRSetEvent ChangeRRSetEvent
	if err := json.Unmarshal(event.Payload, &changeRRSetEvent); err != nil {
		return err
	}

	change, err := p.repository.GetZoneRepository().GetChange(ctx, changeRRSetEvent.ChangeID)
	if err != nil {
		return err
	}

	tx := p.store.ZoneTxn(ctx, changeRRSetEvent.ZoneName)

	for _, changeAction := range change.Actions {
		if err = processChangeAction(tx, changeAction); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	err = p.repository.GetZoneRepository().UpdateChangeStatus(ctx, change.ID, model.ChangeStatusDone)
	if err != nil {
		return err
	}

	return nil
}

func processChangeAction(tx dnsstore.ZoneTransaction, changeAction model.ChangeAction) error {
	if changeAction.ActionType == model.ChangeActionTypeDelete {
		tx.DeleteRRSet(changeAction.ResourceRecordSet.Name, string(changeAction.ResourceRecordSet.Type))
		return nil
	}

	rrset, err := dns.ParseRRs(changeAction.ResourceRecordSet)
	if err != nil {
		return err
	}

	tx.PutRRSet(changeAction.ResourceRecordSet.Name, string(changeAction.ResourceRecordSet.Type), rrset)

	return nil
}
