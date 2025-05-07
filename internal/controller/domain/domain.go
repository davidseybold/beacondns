package domain

import (
	"time"

	"github.com/google/uuid"
)

type ZoneChangeAction string
type RRSetChangeAction string
type ChangeSyncStatus string
type RRType string

const (
	ZoneChangeActionCreateZone ZoneChangeAction = "CREATE_ZONE"
	ZoneChangeActionDeleteZone ZoneChangeAction = "DELETE_ZONE"
	ZoneChangeActionUpdateZone ZoneChangeAction = "UPDATE_ZONE"

	RRSetChangeActionCreate RRSetChangeAction = "CREATE"
	RRSetChangeActionUpsert RRSetChangeAction = "UPSERT"
	RRSetChangeActionDelete RRSetChangeAction = "DELETE"

	ChangeSyncStatusPending ChangeSyncStatus = "PENDING"
	ChangeSyncStatusInSync  ChangeSyncStatus = "INSYNC"
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

type ZoneInfo struct {
	Zone          Zone
	DelegationSet *DelegationSet
}

type CreateZoneResult struct {
	ZoneInfo
	ChangeInfo ChangeInfo
}

type Zone struct {
	ID        uuid.UUID
	Name      string
	IsPrivate bool
}

type ResourceRecordSet struct {
	ID              uuid.UUID
	Name            string
	Type            RRType
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
	Action            RRSetChangeAction
	ResourceRecordSet ResourceRecordSet
}

type ZoneChange struct {
	ID      uuid.UUID
	ZoneID  uuid.UUID
	Action  ZoneChangeAction
	Changes []ResourceRecordSetChange
}

type ZoneChangeSync struct {
	ZoneChangeID uuid.UUID
	NameServerID uuid.UUID
	Status       ChangeSyncStatus
	SyncedAt     *time.Time
}

type ChangeInfo struct {
	ID          uuid.UUID
	Status      ChangeSyncStatus
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
