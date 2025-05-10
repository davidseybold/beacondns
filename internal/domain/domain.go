package domain

import (
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
	TTL             uint             `json:"ttl"`
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
