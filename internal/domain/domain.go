package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ZoneChangeAction string
type RRSetChangeAction string
type RRType string
type ChangeType string

const (
	ZoneChangeActionCreate ZoneChangeAction = "CREATE"
	ZoneChangeActionDelete ZoneChangeAction = "DELETE"
	ZoneChangeActionUpdate ZoneChangeAction = "UPDATE"

	RRSetChangeActionCreate RRSetChangeAction = "CREATE"
	RRSetChangeActionUpsert RRSetChangeAction = "UPSERT"
	RRSetChangeActionDelete RRSetChangeAction = "DELETE"

	ChangeTypeZone ChangeType = "ZONE"
)

const (
	RRTypeSOA    RRType = "SOA"
	RRTypeNS     RRType = "NS"
	RRTypeA      RRType = "A"
	RRTypeAAAA   RRType = "AAAA"
	RRTypeCNAME  RRType = "CNAME"
	RRTypeMX     RRType = "MX"
	RRTypeTXT    RRType = "TXT"
	RRTypeSRV    RRType = "SRV"
	RRTypePTR    RRType = "PTR"
	RRTypeCAA    RRType = "CAA"
	RRTypeNAPTR  RRType = "NAPTR"
	RRTypeDS     RRType = "DS"
	RRTypeDNSKEY RRType = "DNSKEY"
	RRTypeRRSIG  RRType = "RRSIG"
	RRTypeNSEC   RRType = "NSEC"
	RRTypeTLSA   RRType = "TLSA"
	RRTypeSPF    RRType = "SPF"
)

type ResourceRecordSet struct {
	Name            string           `json:"name"`
	Type            RRType           `json:"type"`
	TTL             uint32           `json:"ttl"`
	ResourceRecords []ResourceRecord `json:"resourceRecords"`
}

type ResourceRecord struct {
	Value string `json:"value"`
}

type ResourceRecordSetChange struct {
	Action            RRSetChangeAction `json:"action"`
	ResourceRecordSet ResourceRecordSet `json:"resourceRecordSet"`
}

type ZoneChange struct {
	ZoneName  string                    `json:"zoneName"`
	Action    ZoneChangeAction          `json:"action"`
	Changes   []ResourceRecordSetChange `json:"changes"`
	Submitted *time.Time                `json:"submitted,omitempty"`
}

type Change struct {
	ID          uuid.UUID   `json:"id"`
	Type        ChangeType  `json:"type"`
	ZoneChange  *ZoneChange `json:"zoneChange,omitempty"`
	SubmittedAt *time.Time  `json:"submittedAt,omitempty"`
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

func NewZoneChange(zoneName string, action ZoneChangeAction, changes []ResourceRecordSetChange) ZoneChange {
	return ZoneChange{
		ZoneName: zoneName,
		Action:   action,
		Changes:  changes,
	}
}

func NewResourceRecordSetChange(action RRSetChangeAction, resourceRecordSet ResourceRecordSet) ResourceRecordSetChange {
	return ResourceRecordSetChange{
		Action:            action,
		ResourceRecordSet: resourceRecordSet,
	}
}

func NewChange(t ChangeType) Change {
	return Change{
		ID:   uuid.New(),
		Type: t,
	}
}

func NewChangeWithZoneChange(zoneChange ZoneChange) Change {
	ch := NewChange(ChangeTypeZone)
	ch.ZoneChange = &zoneChange
	return ch
}
