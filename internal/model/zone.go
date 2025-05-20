package model

import (
	"fmt"

	"github.com/google/uuid"
)

type RRType string

const (
	RRTypeSOA   RRType = "SOA"
	RRTypeNS    RRType = "NS"
	RRTypeA     RRType = "A"
	RRTypeAAAA  RRType = "AAAA"
	RRTypeCNAME RRType = "CNAME"
	RRTypeMX    RRType = "MX"
	RRTypeTXT   RRType = "TXT"
	RRTypeSRV   RRType = "SRV"
	RRTypePTR   RRType = "PTR"
)

type ZoneInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Zone struct {
	ZoneInfo
	ResourceRecordSets []ResourceRecordSet `json:"resourceRecordSets"`
}

func NewZone(name string) Zone {
	return Zone{
		ZoneInfo: ZoneInfo{
			ID:   uuid.New(),
			Name: name,
		},
	}
}

type ResourceRecordSet struct {
	ID              uuid.UUID        `json:"id"`
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
		ID:   uuid.New(),
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
		ID:              uuid.New(),
		Name:            zoneName,
		Type:            RRTypeNS,
		TTL:             ttl,
		ResourceRecords: resourceRecords,
	}
}
