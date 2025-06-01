package zone

import (
	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/model"
)

const (
	EventTypeCreateZone  = "zone.create"
	EventTypeDeleteZone  = "zone.delete"
	EventTypeChangeRRSet = "zone.changeRRSet"
)

type CreateZoneEvent struct {
	ZoneName string    `json:"zoneName"`
	ChangeID uuid.UUID `json:"changeId"`
}

func NewCreateZoneEvent(zoneName string, changeID uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeCreateZone, &CreateZoneEvent{
		ZoneName: zoneName,
		ChangeID: changeID,
	})
}

type ChangeRRSetEvent struct {
	ZoneName string    `json:"zoneName"`
	ChangeID uuid.UUID `json:"changeId"`
}

func NewChangeRRSetEvent(zoneName string, changeID uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeChangeRRSet, &ChangeRRSetEvent{
		ZoneName: zoneName,
		ChangeID: changeID,
	})
}

type DeleteZoneEvent struct {
	ZoneName string `json:"zoneName"`
}

func NewDeleteZoneEvent(zoneName string) *model.Event {
	return model.NewEvent(EventTypeDeleteZone, &DeleteZoneEvent{
		ZoneName: zoneName,
	})
}
