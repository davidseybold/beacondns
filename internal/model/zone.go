package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RRType string

const (
	RRTypeSOA   RRType = "SOA"
	RRTypeNS    RRType = "NS"
	RRTypeA     RRType = "A"
	RRTypeAAAA  RRType = "AAAA"
	RRTypeCNAME RRType = "CNAME"
	RRTypeCAA   RRType = "CAA"
	RRTypePTR   RRType = "PTR"
	RRTypeSSHFP RRType = "SSHFP"
	RRTypeSVCB  RRType = "SVCB"
	RRTypeTLSA  RRType = "TLSA"
	RRTypeSRV   RRType = "SRV"
	RRTypeTXT   RRType = "TXT"
	RRTypeNAPTR RRType = "NAPTR"
	RRTypeDS    RRType = "DS"
	RRTypeHTTPS RRType = "HTTPS"
	RRTypeMX    RRType = "MX"
)

var SupportedRRTypes = map[RRType]struct{}{
	RRTypeSOA:   {},
	RRTypeNS:    {},
	RRTypeA:     {},
	RRTypeAAAA:  {},
	RRTypeCNAME: {},
	RRTypeCAA:   {},
	RRTypePTR:   {},
	RRTypeSSHFP: {},
	RRTypeSVCB:  {},
	RRTypeTLSA:  {},
	RRTypeSRV:   {},
	RRTypeTXT:   {},
	RRTypeNAPTR: {},
	RRTypeDS:    {},
	RRTypeHTTPS: {},
	RRTypeMX:    {},
}

type ZoneInfo struct {
	ID                     uuid.UUID `json:"id"`
	Name                   string    `json:"name"`
	ResourceRecordSetCount int       `json:"resourceRecordSetCount"`
}

type Zone struct {
	ID                 uuid.UUID           `json:"id"`
	Name               string              `json:"name"`
	ResourceRecordSets []ResourceRecordSet `json:"resourceRecordSets"`
}

func NewZone(name string) *Zone {
	return &Zone{
		ID:   uuid.New(),
		Name: name,
	}
}

type ResourceRecordSet struct {
	Name            string           `json:"name"`
	Type            RRType           `json:"type"`
	TTL             uint32           `json:"ttl"`
	ResourceRecords []ResourceRecord `json:"resourceRecords"`
}

type ResourceRecord struct {
	Value string `json:"value"`
}

func NewSOA(
	zoneName string,
	ttl uint32,
	primaryNS string,
	hostmasterEmail string,
	soaSerial uint,
	soaRefresh uint,
	soaRetry uint,
	soaExpire uint,
	soaMinimum uint,
) ResourceRecordSet {
	return ResourceRecordSet{
		Name: zoneName,
		Type: RRTypeSOA,
		TTL:  ttl,
		ResourceRecords: []ResourceRecord{
			{
				Value: fmt.Sprintf(
					"%s %s %d %d %d %d %d",
					primaryNS,
					hostmasterEmail,
					soaSerial,
					soaRefresh,
					soaRetry,
					soaExpire,
					soaMinimum,
				),
			},
		},
	}
}

func NewNS(zoneName string, ttl uint32, nameServerNames []string) ResourceRecordSet {
	resourceRecords := make([]ResourceRecord, len(nameServerNames))
	for i, nameServer := range nameServerNames {
		resourceRecords[i] = ResourceRecord{
			Value: nameServer,
		}
	}

	return ResourceRecordSet{
		Name:            zoneName,
		Type:            RRTypeNS,
		TTL:             ttl,
		ResourceRecords: resourceRecords,
	}
}

type ChangeStatus string

const (
	ChangeStatusPending ChangeStatus = "PENDING"
	ChangeStatusDone    ChangeStatus = "DONE"
)

type ChangeActionType string

const (
	ChangeActionTypeUpsert ChangeActionType = "UPSERT"
	ChangeActionTypeDelete ChangeActionType = "DELETE"
)

type Change struct {
	ID          uuid.UUID      `json:"id"`
	ZoneID      uuid.UUID      `json:"zoneID"`
	Actions     []ChangeAction `json:"actions"`
	Status      ChangeStatus   `json:"status"`
	SubmittedAt *time.Time     `json:"submittedAt,omitempty"`
}

func NewChange(zoneID uuid.UUID, status ChangeStatus, actions []ChangeAction) Change {
	return Change{
		ID:      uuid.New(),
		ZoneID:  zoneID,
		Status:  status,
		Actions: actions,
	}
}

type ChangeAction struct {
	ActionType        ChangeActionType   `json:"actionType"`
	ResourceRecordSet *ResourceRecordSet `json:"resourceRecordSet"`
}

func NewChangeAction(actionType ChangeActionType, resourceRecordSet *ResourceRecordSet) ChangeAction {
	return ChangeAction{
		ActionType:        actionType,
		ResourceRecordSet: resourceRecordSet,
	}
}
