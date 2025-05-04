package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	ZoneChangeActionCreateZone = "CREATE_ZONE"
	ZoneChangeActionDeleteZone = "DELETE_ZONE"
	ZoneChangeActionUpdateZone = "UPDATE_ZONE"

	RRSetChangeActionCreate = "CREATE"
	RRSetChangeActionUpsert = "UPSERT"
	RRSetChangeActionDelete = "DELETE"

	ChangeSyncStatusPending = "PENDING"
	ChangeSyncStatusInSync  = "INSYNC"
)

const (
	RRTypeSOA    = "SOA"
	RRTypeNS     = "NS"
	RRTypeA      = "A"
	RRTypeAAAA   = "AAAA"
	RRTypeCNAME  = "CNAME"
	RRTypeMX     = "MX"
	RRTypeTXT    = "TXT"
	RRTypeSRV    = "SRV"
	RRTypePTR    = "PTR"
	RRTypeCAA    = "CAA"
	RRTypeNAPTR  = "NAPTR"
	RRTypeDS     = "DS"
	RRTypeDNSKEY = "DNSKEY"
	RRTypeRRSIG  = "RRSIG"
	RRTypeNSSEC  = "NSEC"
	RRTypeTLSA   = "TLSA"
	RRTypeSPF    = "SPF"
)

type ZoneInfo struct {
	Zone          Zone
	DelegationSet DelegationSet
}

type CreateZoneResult struct {
	ZoneInfo
	ChangeInfo ChangeInfo
}

type Zone struct {
	ID   uuid.UUID
	Name string
}

type ResourceRecordSet struct {
	ID              uuid.UUID
	Name            string
	Type            string
	TTL             uint
	ResourceRecords []ResourceRecord
}

type ResourceRecord struct {
	Value string
}

type ChangeBatch struct {
	Changes []ResourceRecordSetChange
}

type ResourceRecordSetChange struct {
	Action            string
	ResourceRecordSet ResourceRecordSet
}

type ZoneChange struct {
	ID      uuid.UUID
	ZoneID  uuid.UUID
	Action  string
	Changes []ResourceRecordSetChange
}

type ZoneChangeSync struct {
	ZoneChangeID uuid.UUID
	NameServerID uuid.UUID
	Status       string
	SyncedAt     *time.Time
}

type ChangeInfo struct {
	ID          uuid.UUID
	Status      string
	SubmittedAt time.Time
}

type NameServer struct {
	ID        uuid.UUID
	Name      string
	IPAddress string
	RouteKey  string
}

type DelegationSet struct {
	ID          uuid.UUID
	NameServers []NameServer
}

type OutboxMessage struct {
	ID       uuid.UUID
	Payload  []byte
	RouteKey string
}
