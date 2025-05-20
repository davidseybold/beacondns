package model

import (
	"time"

	"github.com/google/uuid"
)

type ChangeType string

const (
	ChangeTypeZone ChangeType = "ZONE"
)

type ChangeStatus string
type ChangeTargetStatus string

const (
	ChangeStatusPending ChangeStatus = "PENDING"
	ChangeStatusInSync  ChangeStatus = "INSYNC"
)

type ZoneChangeAction string
type RRSetChangeAction string

const (
	ZoneChangeActionCreate ZoneChangeAction = "CREATE"
	ZoneChangeActionDelete ZoneChangeAction = "DELETE"
	ZoneChangeActionUpdate ZoneChangeAction = "UPDATE"

	RRSetChangeActionCreate RRSetChangeAction = "CREATE"
	RRSetChangeActionUpsert RRSetChangeAction = "UPSERT"
	RRSetChangeActionDelete RRSetChangeAction = "DELETE"
)

type Change struct {
	ID         uuid.UUID    `json:"id"`
	Type       ChangeType   `json:"type"`
	ZoneChange *ZoneChange  `json:"zoneChange,omitempty"`
	Status     ChangeStatus `json:"status"`

	SubmittedAt *time.Time `json:"submittedAt,omitempty"`
}

func NewChange(t ChangeType, status ChangeStatus) Change {
	return Change{
		ID:     uuid.New(),
		Type:   t,
		Status: status,
	}
}

func NewChangeWithZoneChange(zoneChange ZoneChange, status ChangeStatus) Change {
	ch := NewChange(ChangeTypeZone, status)
	ch.ZoneChange = &zoneChange
	return ch
}

type ZoneChange struct {
	ZoneName    string                    `json:"zoneName"`
	Action      ZoneChangeAction          `json:"action"`
	Changes     []ResourceRecordSetChange `json:"changes"`
	SubmittedAt *time.Time                `json:"submittedAt,omitempty"`
}

func NewZoneChange(zoneName string, action ZoneChangeAction, changes []ResourceRecordSetChange) ZoneChange {
	return ZoneChange{
		ZoneName: zoneName,
		Action:   action,
		Changes:  changes,
	}
}

type ResourceRecordSetChange struct {
	Action            RRSetChangeAction `json:"action"`
	ResourceRecordSet ResourceRecordSet `json:"resourceRecordSet"`
}

func NewResourceRecordSetChange(action RRSetChangeAction, resourceRecordSet ResourceRecordSet) ResourceRecordSetChange {
	return ResourceRecordSetChange{
		Action:            action,
		ResourceRecordSet: resourceRecordSet,
	}
}
