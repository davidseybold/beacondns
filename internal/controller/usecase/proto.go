package usecase

import (
	"github.com/davidseybold/beacondns/internal/controller/domain"
	beacondnspb "github.com/davidseybold/beacondns/internal/libs/gen/proto/beacondns/v1"
)

var domainZoneChangeActionToProto = map[domain.ZoneChangeAction]beacondnspb.ZoneChangeAction{
	domain.ZoneChangeActionCreateZone: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_CREATE_ZONE,
	domain.ZoneChangeActionUpdateZone: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UPDATE_ZONE,
	domain.ZoneChangeActionDeleteZone: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_DELETE_ZONE,
}

var domainRRSetChangeActionToProto = map[domain.RRSetChangeAction]beacondnspb.RRSetChangeAction{
	domain.RRSetChangeActionCreate: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE,
	domain.RRSetChangeActionUpsert: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UPSERT,
	domain.RRSetChangeActionDelete: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_DELETE,
}

var domainRRTypeToProto = map[domain.RRType]beacondnspb.RRType{
	domain.RRTypeSOA:    beacondnspb.RRType_RR_TYPE_SOA,
	domain.RRTypeNS:     beacondnspb.RRType_RR_TYPE_NS,
	domain.RRTypeA:      beacondnspb.RRType_RR_TYPE_A,
	domain.RRTypeAAAA:   beacondnspb.RRType_RR_TYPE_AAAA,
	domain.RRTypeCNAME:  beacondnspb.RRType_RR_TYPE_CNAME,
	domain.RRTypeMX:     beacondnspb.RRType_RR_TYPE_MX,
	domain.RRTypeTXT:    beacondnspb.RRType_RR_TYPE_TXT,
	domain.RRTypeSRV:    beacondnspb.RRType_RR_TYPE_SRV,
	domain.RRTypePTR:    beacondnspb.RRType_RR_TYPE_PTR,
	domain.RRTypeCAA:    beacondnspb.RRType_RR_TYPE_CAA,
	domain.RRTypeNAPTR:  beacondnspb.RRType_RR_TYPE_NAPTR,
	domain.RRTypeDS:     beacondnspb.RRType_RR_TYPE_DS,
	domain.RRTypeDNSKEY: beacondnspb.RRType_RR_TYPE_DNSKEY,
	domain.RRTypeRRSIG:  beacondnspb.RRType_RR_TYPE_RRSIG,
	domain.RRTypeNSEC:   beacondnspb.RRType_RR_TYPE_NSEC,
	domain.RRTypeTLSA:   beacondnspb.RRType_RR_TYPE_TLSA,
	domain.RRTypeSPF:    beacondnspb.RRType_RR_TYPE_SPF,
}

func newZoneChangeEvent(zoneName string, action domain.ZoneChangeAction, changes []domain.ResourceRecordSetChange) *beacondnspb.ZoneChangeEvent {
	pbZoneAction, ok := domainZoneChangeActionToProto[action]
	if !ok {
		pbZoneAction = beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UNSPECIFIED
	}

	pbChanges := make([]*beacondnspb.ResourceRecordSetChange, len(changes))
	for i, change := range changes {
		pbAction, ok := domainRRSetChangeActionToProto[change.Action]
		if !ok {
			pbAction = beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UNSPECIFIED
		}

		pbType, ok := domainRRTypeToProto[change.ResourceRecordSet.Type]
		if !ok {
			pbType = beacondnspb.RRType_RR_TYPE_UNSPECIFIED
		}

		records := make([]*beacondnspb.ResourceRecord, len(change.ResourceRecordSet.ResourceRecords))
		for j, rr := range change.ResourceRecordSet.ResourceRecords {
			records[j] = &beacondnspb.ResourceRecord{
				Value: rr.Value,
			}
		}

		pbChanges[i] = &beacondnspb.ResourceRecordSetChange{
			Action: pbAction,
			ResourceRecordSet: &beacondnspb.ResourceRecordSet{
				Name:            change.ResourceRecordSet.Name,
				Type:            pbType,
				Ttl:             uint32(change.ResourceRecordSet.TTL),
				ResourceRecords: records,
			},
		}
	}

	return &beacondnspb.ZoneChangeEvent{
		ZoneName: zoneName,
		Action:   pbZoneAction,
		Changes:  pbChanges,
	}
}
