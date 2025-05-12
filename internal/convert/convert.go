package convert

import (
	beacondomain "github.com/davidseybold/beacondns/internal/domain"
	beacondnspb "github.com/davidseybold/beacondns/internal/libs/gen/proto/beacondns/v1"
)

func DomainChangeToProto(domain *beacondomain.Change) *beacondnspb.Change {
	change := &beacondnspb.Change{
		Id:   domain.ID.String(),
		Type: DomainChangeTypeToProto(domain.Type),
	}

	if domain.Type == beacondomain.ChangeTypeZone {
		change.ZoneChange = DomainZoneChangeToProto(domain.ZoneChange)
	}

	return change
}

func DomainChangeTypeToProto(changeType beacondomain.ChangeType) beacondnspb.ChangeType {
	switch changeType {
	case beacondomain.ChangeTypeZone:
		return beacondnspb.ChangeType_CHANGE_TYPE_ZONE
	default:
		return beacondnspb.ChangeType_CHANGE_TYPE_UNSPECIFIED
	}
}

func DomainZoneChangeToProto(change *beacondomain.ZoneChange) *beacondnspb.ZoneChange {
	rrChanges := make([]*beacondnspb.ResourceRecordSetChange, len(change.Changes))
	for i, change := range change.Changes {
		rrChanges[i] = DomainResourceRecordSetChangeToProto(&change)
	}

	return &beacondnspb.ZoneChange{
		ZoneName: change.ZoneName,
		Action:   DomainZoneChangeActionToProto(change.Action),
		Changes:  rrChanges,
	}
}

func DomainZoneChangeActionToProto(action beacondomain.ZoneChangeAction) beacondnspb.ZoneChangeAction {
	switch action {
	case beacondomain.ZoneChangeActionCreate:
		return beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_CREATE
	case beacondomain.ZoneChangeActionDelete:
		return beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_DELETE
	case beacondomain.ZoneChangeActionUpdate:
		return beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UPDATE
	default:
		return beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UNSPECIFIED
	}
}

func DomainResourceRecordSetChangeToProto(
	change *beacondomain.ResourceRecordSetChange,
) *beacondnspb.ResourceRecordSetChange {
	return &beacondnspb.ResourceRecordSetChange{
		Action:            DomainResourceRecordSetChangeActionToProto(change.Action),
		ResourceRecordSet: DomainResourceRecordSetToProto(&change.ResourceRecordSet),
	}
}

func DomainResourceRecordSetChangeActionToProto(action beacondomain.RRSetChangeAction) beacondnspb.RRSetChangeAction {
	switch action {
	case beacondomain.RRSetChangeActionCreate:
		return beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE
	case beacondomain.RRSetChangeActionDelete:
		return beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_DELETE
	case beacondomain.RRSetChangeActionUpsert:
		return beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UPSERT
	default:
		return beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UNSPECIFIED
	}
}

func DomainResourceRecordSetToProto(rrset *beacondomain.ResourceRecordSet) *beacondnspb.ResourceRecordSet {
	resourceRecords := make([]*beacondnspb.ResourceRecord, len(rrset.ResourceRecords))
	for i, resourceRecord := range rrset.ResourceRecords {
		resourceRecords[i] = DomainResourceRecordToProto(&resourceRecord)
	}

	return &beacondnspb.ResourceRecordSet{
		Name:            rrset.Name,
		Type:            DomainRRTypeToProto(rrset.Type),
		Ttl:             rrset.TTL,
		ResourceRecords: resourceRecords,
	}
}

func DomainRRTypeToProto(rrType beacondomain.RRType) beacondnspb.RRType {
	switch rrType {
	case beacondomain.RRTypeA:
		return beacondnspb.RRType_RR_TYPE_A
	case beacondomain.RRTypeAAAA:
		return beacondnspb.RRType_RR_TYPE_AAAA
	case beacondomain.RRTypeCNAME:
		return beacondnspb.RRType_RR_TYPE_CNAME
	case beacondomain.RRTypeMX:
		return beacondnspb.RRType_RR_TYPE_MX
	case beacondomain.RRTypeNS:
		return beacondnspb.RRType_RR_TYPE_NS
	case beacondomain.RRTypePTR:
		return beacondnspb.RRType_RR_TYPE_PTR
	case beacondomain.RRTypeSOA:
		return beacondnspb.RRType_RR_TYPE_SOA
	case beacondomain.RRTypeTXT:
		return beacondnspb.RRType_RR_TYPE_TXT
	case beacondomain.RRTypeSRV:
		return beacondnspb.RRType_RR_TYPE_SRV
	case beacondomain.RRTypeSPF:
		return beacondnspb.RRType_RR_TYPE_SPF
	case beacondomain.RRTypeCAA:
		return beacondnspb.RRType_RR_TYPE_CAA
	case beacondomain.RRTypeNAPTR:
		return beacondnspb.RRType_RR_TYPE_NAPTR
	case beacondomain.RRTypeDS:
		return beacondnspb.RRType_RR_TYPE_DS
	case beacondomain.RRTypeNSEC:
		return beacondnspb.RRType_RR_TYPE_NSEC
	case beacondomain.RRTypeDNSKEY:
		return beacondnspb.RRType_RR_TYPE_DNSKEY
	case beacondomain.RRTypeRRSIG:
		return beacondnspb.RRType_RR_TYPE_RRSIG
	case beacondomain.RRTypeTLSA:
		return beacondnspb.RRType_RR_TYPE_TLSA
	default:
		return beacondnspb.RRType_RR_TYPE_UNSPECIFIED
	}
}

func DomainResourceRecordToProto(resourceRecord *beacondomain.ResourceRecord) *beacondnspb.ResourceRecord {
	return &beacondnspb.ResourceRecord{
		Value: resourceRecord.Value,
	}
}
