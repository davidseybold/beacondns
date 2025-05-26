package model

import (
	"time"

	"github.com/google/uuid"
)

type ChangeType string

const (
	ChangeTypeZone           ChangeType = "ZONE"
	ChangeTypeResponsePolicy ChangeType = "RESPONSE_POLICY"
)

type ChangeStatus string

const (
	ChangeStatusPending ChangeStatus = "PENDING"
	ChangeStatusDone    ChangeStatus = "DONE"
)

type ZoneChangeAction string
type RRSetChangeAction string

type ResponsePolicyChangeAction string
type ResponsePolicyRuleChangeAction string

const (
	ZoneChangeActionCreate ZoneChangeAction = "CREATE"
	ZoneChangeActionDelete ZoneChangeAction = "DELETE"
	ZoneChangeActionUpdate ZoneChangeAction = "UPDATE"

	RRSetChangeActionUpsert RRSetChangeAction = "UPSERT"
	RRSetChangeActionDelete RRSetChangeAction = "DELETE"

	ResponsePolicyChangeActionUpsert ResponsePolicyChangeAction = "UPSERT"
	ResponsePolicyChangeActionDelete ResponsePolicyChangeAction = "DELETE"

	ResponsePolicyRuleChangeActionUpsert ResponsePolicyRuleChangeAction = "UPSERT"
	ResponsePolicyRuleChangeActionDelete ResponsePolicyRuleChangeAction = "DELETE"
)

type Change struct {
	ID                   uuid.UUID             `json:"id"`
	Type                 ChangeType            `json:"type"`
	ZoneChange           *ZoneChange           `json:"zoneChange,omitempty"`
	ResponsePolicyChange *ResponsePolicyChange `json:"responsePolicyChange,omitempty"`
	Status               ChangeStatus          `json:"status"`

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
	ZoneName string                    `json:"zoneName"`
	Action   ZoneChangeAction          `json:"action"`
	Changes  []ResourceRecordSetChange `json:"changes"`
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

type ResponsePolicyChange struct {
	ResponsePolicyID uuid.UUID                  `json:"responsePolicyID"`
	Action           ResponsePolicyChangeAction `json:"action"`
	Changes          []ResourceRecordSetChange  `json:"changes"`
}
