package convert

import (
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	beacondnspb "github.com/davidseybold/beacondns/internal/gen/proto/beacondns/v1"
	"github.com/davidseybold/beacondns/internal/model"
)

func ChangeToProto(ch *model.Change) *beacondnspb.Change {
	if ch == nil {
		return nil
	}

	var submittedAt *timestamppb.Timestamp
	if ch.SubmittedAt != nil {
		submittedAt = timestamppb.New(*ch.SubmittedAt)
	}

	return &beacondnspb.Change{
		Id:          ch.ID.String(),
		Type:        ChangeTypeToProto(ch.Type),
		ZoneChange:  ZoneChangeToProto(ch.ZoneChange),
		SubmittedAt: submittedAt,
	}
}

func ChangeFromProto(ch *beacondnspb.Change) *model.Change {
	if ch == nil {
		return nil
	}

	submittedAt := ch.GetSubmittedAt().AsTime()
	return &model.Change{
		ID:          uuid.MustParse(ch.GetId()),
		Type:        ChangeTypeFromProto(ch.GetType()),
		SubmittedAt: &submittedAt,
		ZoneChange:  ZoneChangeFromProto(ch.GetZoneChange()),
	}
}

func ChangeTypeToProto(chType model.ChangeType) beacondnspb.ChangeType {
	switch chType {
	case model.ChangeTypeZone:
		return beacondnspb.ChangeType_CHANGE_TYPE_ZONE
	default:
		return beacondnspb.ChangeType_CHANGE_TYPE_UNSPECIFIED
	}
}

func ChangeTypeFromProto(chType beacondnspb.ChangeType) model.ChangeType {
	switch chType {
	case beacondnspb.ChangeType_CHANGE_TYPE_ZONE:
		return model.ChangeTypeZone
	case beacondnspb.ChangeType_CHANGE_TYPE_UNSPECIFIED:
		return model.ChangeTypeZone
	default:
		return model.ChangeTypeZone
	}
}

func ZoneChangeToProto(ch *model.ZoneChange) *beacondnspb.ZoneChange {
	if ch == nil {
		return nil
	}

	changes := make([]*beacondnspb.ResourceRecordSetChange, len(ch.Changes))
	for i, change := range ch.Changes {
		changes[i] = ResourceRecordSetChangeToProto(&change)
	}

	return &beacondnspb.ZoneChange{
		ZoneName: ch.ZoneName,
		Action:   ZoneChangeActionToProto(ch.Action),
		Changes:  changes,
	}
}

func ZoneChangeFromProto(ch *beacondnspb.ZoneChange) *model.ZoneChange {
	if ch == nil {
		return nil
	}

	changes := make([]model.ResourceRecordSetChange, len(ch.GetChanges()))
	for i, change := range ch.GetChanges() {
		changes[i] = *ResourceRecordSetChangeFromProto(change)
	}

	return &model.ZoneChange{
		ZoneName: ch.GetZoneName(),
		Action:   ZoneChangeActionFromProto(ch.GetAction()),
		Changes:  changes,
	}
}

func ZoneChangeActionToProto(chAction model.ZoneChangeAction) beacondnspb.ZoneChangeAction {
	switch chAction {
	case model.ZoneChangeActionCreate:
		return beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_CREATE
	case model.ZoneChangeActionUpdate:
		return beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UPDATE
	case model.ZoneChangeActionDelete:
		return beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_DELETE
	default:
		return beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UNSPECIFIED
	}
}

func ZoneChangeActionFromProto(chAction beacondnspb.ZoneChangeAction) model.ZoneChangeAction {
	switch chAction {
	case beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_CREATE:
		return model.ZoneChangeActionCreate
	case beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UPDATE:
		return model.ZoneChangeActionUpdate
	case beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_DELETE:
		return model.ZoneChangeActionDelete
	case beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UNSPECIFIED:
		return model.ZoneChangeActionCreate
	default:
		return model.ZoneChangeActionCreate
	}
}

func ResourceRecordSetChangeToProto(ch *model.ResourceRecordSetChange) *beacondnspb.ResourceRecordSetChange {
	if ch == nil {
		return nil
	}

	return &beacondnspb.ResourceRecordSetChange{
		Action:            RRSetChangeActionToProto(ch.Action),
		ResourceRecordSet: ResourceRecordSetToProto(&ch.ResourceRecordSet),
	}
}

func ResourceRecordSetChangeFromProto(ch *beacondnspb.ResourceRecordSetChange) *model.ResourceRecordSetChange {
	if ch == nil {
		return nil
	}

	return &model.ResourceRecordSetChange{
		Action:            RRSetChangeActionFromProto(ch.GetAction()),
		ResourceRecordSet: *ResourceRecordSetFromProto(ch.GetResourceRecordSet()),
	}
}

func RRSetChangeActionToProto(chAction model.RRSetChangeAction) beacondnspb.RRSetChangeAction {
	switch chAction {
	case model.RRSetChangeActionCreate:
		return beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE
	case model.RRSetChangeActionUpsert:
		return beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UPSERT
	case model.RRSetChangeActionDelete:
		return beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_DELETE
	default:
		return beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UNSPECIFIED
	}
}

func RRSetChangeActionFromProto(chAction beacondnspb.RRSetChangeAction) model.RRSetChangeAction {
	switch chAction {
	case beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE:
		return model.RRSetChangeActionCreate
	case beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UPSERT:
		return model.RRSetChangeActionUpsert
	case beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_DELETE:
		return model.RRSetChangeActionDelete
	case beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UNSPECIFIED:
		return model.RRSetChangeActionCreate
	default:
		return model.RRSetChangeActionCreate
	}
}

func ResourceRecordSetToProto(rrset *model.ResourceRecordSet) *beacondnspb.ResourceRecordSet {
	if rrset == nil {
		return nil
	}

	records := make([]*beacondnspb.ResourceRecord, len(rrset.ResourceRecords))
	for i, record := range rrset.ResourceRecords {
		records[i] = ResourceRecordToProto(&record)
	}

	return &beacondnspb.ResourceRecordSet{
		Name:            rrset.Name,
		Type:            RRTypeToProto(rrset.Type),
		Ttl:             rrset.TTL,
		ResourceRecords: records,
	}
}

func ResourceRecordSetFromProto(rrset *beacondnspb.ResourceRecordSet) *model.ResourceRecordSet {
	if rrset == nil {
		return nil
	}

	records := make([]model.ResourceRecord, len(rrset.GetResourceRecords()))
	for i, record := range rrset.GetResourceRecords() {
		records[i] = *ResourceRecordFromProto(record)
	}

	return &model.ResourceRecordSet{
		Name:            rrset.GetName(),
		Type:            RRTypeFromProto(rrset.GetType()),
		TTL:             rrset.GetTtl(),
		ResourceRecords: records,
	}
}

func ResourceRecordToProto(rr *model.ResourceRecord) *beacondnspb.ResourceRecord {
	if rr == nil {
		return nil
	}

	return &beacondnspb.ResourceRecord{
		Value: rr.Value,
	}
}

func ResourceRecordFromProto(rr *beacondnspb.ResourceRecord) *model.ResourceRecord {
	if rr == nil {
		return nil
	}

	return &model.ResourceRecord{
		Value: rr.GetValue(),
	}
}

func RRTypeToProto(chType model.RRType) beacondnspb.RRType {
	switch chType {
	case model.RRTypeA:
		return beacondnspb.RRType_RR_TYPE_A
	case model.RRTypeAAAA:
		return beacondnspb.RRType_RR_TYPE_AAAA
	case model.RRTypeCNAME:
		return beacondnspb.RRType_RR_TYPE_CNAME
	case model.RRTypeMX:
		return beacondnspb.RRType_RR_TYPE_MX
	case model.RRTypeNS:
		return beacondnspb.RRType_RR_TYPE_NS
	case model.RRTypePTR:
		return beacondnspb.RRType_RR_TYPE_PTR
	case model.RRTypeSOA:
		return beacondnspb.RRType_RR_TYPE_SOA
	case model.RRTypeSRV:
		return beacondnspb.RRType_RR_TYPE_SRV
	case model.RRTypeTXT:
		return beacondnspb.RRType_RR_TYPE_TXT
	default:
		return beacondnspb.RRType_RR_TYPE_UNSPECIFIED
	}
}

func RRTypeFromProto(chType beacondnspb.RRType) model.RRType {
	switch chType {
	case beacondnspb.RRType_RR_TYPE_A:
		return model.RRTypeA
	case beacondnspb.RRType_RR_TYPE_AAAA:
		return model.RRTypeAAAA
	case beacondnspb.RRType_RR_TYPE_CNAME:
		return model.RRTypeCNAME
	case beacondnspb.RRType_RR_TYPE_MX:
		return model.RRTypeMX
	case beacondnspb.RRType_RR_TYPE_NS:
		return model.RRTypeNS
	case beacondnspb.RRType_RR_TYPE_PTR:
		return model.RRTypePTR
	case beacondnspb.RRType_RR_TYPE_SOA:
		return model.RRTypeSOA
	case beacondnspb.RRType_RR_TYPE_SRV:
		return model.RRTypeSRV
	case beacondnspb.RRType_RR_TYPE_TXT:
		return model.RRTypeTXT
	case beacondnspb.RRType_RR_TYPE_UNSPECIFIED:
		return model.RRTypeA
	default:
		return model.RRTypeA
	}
}
