package convert

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	beacondnspb "github.com/davidseybold/beacondns/internal/gen/proto/beacondns/v1"
	"github.com/davidseybold/beacondns/internal/model"
)

func TestChangeConversions(t *testing.T) {
	tests := []struct {
		name     string
		change   *model.Change
		expected *beacondnspb.Change
	}{
		{
			name:     "nil change",
			change:   nil,
			expected: nil,
		},
		{
			name: "complete change",
			change: &model.Change{
				ID:   uuid.New(),
				Type: model.ChangeTypeZone,
				ZoneChange: &model.ZoneChange{
					ZoneName: "example.com.",
					Action:   model.ZoneChangeActionCreate,
					Changes: []model.ResourceRecordSetChange{
						{
							Action: model.RRSetChangeActionCreate,
							ResourceRecordSet: model.ResourceRecordSet{
								Name: "example.com.",
								Type: model.RRTypeA,
								TTL:  300,
								ResourceRecords: []model.ResourceRecord{
									{Value: "192.0.2.1"},
								},
							},
						},
					},
				},
				SubmittedAt: &time.Time{},
			},
			expected: &beacondnspb.Change{
				Type: beacondnspb.ChangeType_CHANGE_TYPE_ZONE,
				ZoneChange: &beacondnspb.ZoneChange{
					ZoneName: "example.com.",
					Action:   beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_CREATE,
					Changes: []*beacondnspb.ResourceRecordSetChange{
						{
							Action: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE,
							ResourceRecordSet: &beacondnspb.ResourceRecordSet{
								Name: "example.com.",
								Type: beacondnspb.RRType_RR_TYPE_A,
								Ttl:  300,
								ResourceRecords: []*beacondnspb.ResourceRecord{
									{Value: "192.0.2.1"},
								},
							},
						},
					},
				},
				SubmittedAt: timestamppb.New(time.Time{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			proto := ChangeToProto(tt.change)
			if tt.expected == nil {
				assert.Nil(t, proto)
			} else {
				assert.NotNil(t, proto)
				if tt.change != nil {
					assert.Equal(t, tt.change.ID.String(), proto.GetId())
				}
				assert.Equal(t, tt.expected.GetType(), proto.Type)
				assert.Equal(t, tt.expected.GetZoneChange(), proto.ZoneChange)
				assert.Equal(t, tt.expected.GetSubmittedAt(), proto.SubmittedAt)
			}

			// Test FromProto
			if tt.change != nil {
				model := ChangeFromProto(proto)
				assert.NotNil(t, model)
				assert.Equal(t, tt.change.ID, model.ID)
				assert.Equal(t, tt.change.Type, model.Type)
				assert.Equal(t, tt.change.ZoneChange, model.ZoneChange)
				assert.Equal(t, tt.change.SubmittedAt, model.SubmittedAt)
			}
		})
	}
}

func TestChangeTypeConversions(t *testing.T) {
	tests := []struct {
		name  string
		model model.ChangeType
		proto beacondnspb.ChangeType
	}{
		{
			name:  "zone type",
			model: model.ChangeTypeZone,
			proto: beacondnspb.ChangeType_CHANGE_TYPE_ZONE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			assert.Equal(t, tt.proto, ChangeTypeToProto(tt.model))

			// Test FromProto
			assert.Equal(t, tt.model, ChangeTypeFromProto(tt.proto))
		})
	}
}

func TestZoneChangeConversions(t *testing.T) {
	tests := []struct {
		name     string
		change   *model.ZoneChange
		expected *beacondnspb.ZoneChange
	}{
		{
			name:     "nil change",
			change:   nil,
			expected: nil,
		},
		{
			name: "complete change",
			change: &model.ZoneChange{
				ZoneName: "example.com.",
				Action:   model.ZoneChangeActionCreate,
				Changes: []model.ResourceRecordSetChange{
					{
						Action: model.RRSetChangeActionCreate,
						ResourceRecordSet: model.ResourceRecordSet{
							Name: "example.com.",
							Type: model.RRTypeA,
							TTL:  300,
							ResourceRecords: []model.ResourceRecord{
								{Value: "192.0.2.1"},
							},
						},
					},
				},
			},
			expected: &beacondnspb.ZoneChange{
				ZoneName: "example.com.",
				Action:   beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_CREATE,
				Changes: []*beacondnspb.ResourceRecordSetChange{
					{
						Action: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE,
						ResourceRecordSet: &beacondnspb.ResourceRecordSet{
							Name: "example.com.",
							Type: beacondnspb.RRType_RR_TYPE_A,
							Ttl:  300,
							ResourceRecords: []*beacondnspb.ResourceRecord{
								{Value: "192.0.2.1"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			proto := ZoneChangeToProto(tt.change)
			assert.Equal(t, tt.expected, proto)

			// Test FromProto
			if tt.change != nil {
				model := ZoneChangeFromProto(proto)
				assert.Equal(t, tt.change, model)
			}
		})
	}
}

func TestZoneChangeActionConversions(t *testing.T) {
	tests := []struct {
		name  string
		model model.ZoneChangeAction
		proto beacondnspb.ZoneChangeAction
	}{
		{
			name:  "create action",
			model: model.ZoneChangeActionCreate,
			proto: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_CREATE,
		},
		{
			name:  "update action",
			model: model.ZoneChangeActionUpdate,
			proto: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_UPDATE,
		},
		{
			name:  "delete action",
			model: model.ZoneChangeActionDelete,
			proto: beacondnspb.ZoneChangeAction_ZONE_CHANGE_ACTION_DELETE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			assert.Equal(t, tt.proto, ZoneChangeActionToProto(tt.model))

			// Test FromProto
			assert.Equal(t, tt.model, ZoneChangeActionFromProto(tt.proto))
		})
	}
}

func TestResourceRecordSetChangeConversions(t *testing.T) {
	tests := []struct {
		name     string
		change   *model.ResourceRecordSetChange
		expected *beacondnspb.ResourceRecordSetChange
	}{
		{
			name:     "nil change",
			change:   nil,
			expected: nil,
		},
		{
			name: "complete change",
			change: &model.ResourceRecordSetChange{
				Action: model.RRSetChangeActionCreate,
				ResourceRecordSet: model.ResourceRecordSet{
					Name: "example.com.",
					Type: model.RRTypeA,
					TTL:  300,
					ResourceRecords: []model.ResourceRecord{
						{Value: "192.0.2.1"},
					},
				},
			},
			expected: &beacondnspb.ResourceRecordSetChange{
				Action: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE,
				ResourceRecordSet: &beacondnspb.ResourceRecordSet{
					Name: "example.com.",
					Type: beacondnspb.RRType_RR_TYPE_A,
					Ttl:  300,
					ResourceRecords: []*beacondnspb.ResourceRecord{
						{Value: "192.0.2.1"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			proto := ResourceRecordSetChangeToProto(tt.change)
			assert.Equal(t, tt.expected, proto)

			// Test FromProto
			if tt.change != nil {
				model := ResourceRecordSetChangeFromProto(proto)
				assert.Equal(t, tt.change, model)
			}
		})
	}
}

func TestRRSetChangeActionConversions(t *testing.T) {
	tests := []struct {
		name  string
		model model.RRSetChangeAction
		proto beacondnspb.RRSetChangeAction
	}{
		{
			name:  "create action",
			model: model.RRSetChangeActionCreate,
			proto: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_CREATE,
		},
		{
			name:  "upsert action",
			model: model.RRSetChangeActionUpsert,
			proto: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_UPSERT,
		},
		{
			name:  "delete action",
			model: model.RRSetChangeActionDelete,
			proto: beacondnspb.RRSetChangeAction_RR_SET_CHANGE_ACTION_DELETE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			assert.Equal(t, tt.proto, RRSetChangeActionToProto(tt.model))

			// Test FromProto
			assert.Equal(t, tt.model, RRSetChangeActionFromProto(tt.proto))
		})
	}
}

func TestResourceRecordSetConversions(t *testing.T) {
	tests := []struct {
		name     string
		rrset    *model.ResourceRecordSet
		expected *beacondnspb.ResourceRecordSet
	}{
		{
			name:     "nil rrset",
			rrset:    nil,
			expected: nil,
		},
		{
			name: "complete rrset",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: model.RRTypeA,
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "192.0.2.1"},
					{Value: "192.0.2.2"},
				},
			},
			expected: &beacondnspb.ResourceRecordSet{
				Name: "example.com.",
				Type: beacondnspb.RRType_RR_TYPE_A,
				Ttl:  300,
				ResourceRecords: []*beacondnspb.ResourceRecord{
					{Value: "192.0.2.1"},
					{Value: "192.0.2.2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			proto := ResourceRecordSetToProto(tt.rrset)
			assert.Equal(t, tt.expected, proto)

			// Test FromProto
			if tt.rrset != nil {
				model := ResourceRecordSetFromProto(proto)
				assert.Equal(t, tt.rrset, model)
			}
		})
	}
}

func TestResourceRecordConversions(t *testing.T) {
	tests := []struct {
		name     string
		record   *model.ResourceRecord
		expected *beacondnspb.ResourceRecord
	}{
		{
			name:     "nil record",
			record:   nil,
			expected: nil,
		},
		{
			name: "complete record",
			record: &model.ResourceRecord{
				Value: "192.0.2.1",
			},
			expected: &beacondnspb.ResourceRecord{
				Value: "192.0.2.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			proto := ResourceRecordToProto(tt.record)
			assert.Equal(t, tt.expected, proto)

			// Test FromProto
			if tt.record != nil {
				model := ResourceRecordFromProto(proto)
				assert.Equal(t, tt.record, model)
			}
		})
	}
}

func TestRRTypeConversions(t *testing.T) {
	tests := []struct {
		name  string
		model model.RRType
		proto beacondnspb.RRType
	}{
		{
			name:  "A record",
			model: model.RRTypeA,
			proto: beacondnspb.RRType_RR_TYPE_A,
		},
		{
			name:  "AAAA record",
			model: model.RRTypeAAAA,
			proto: beacondnspb.RRType_RR_TYPE_AAAA,
		},
		{
			name:  "CNAME record",
			model: model.RRTypeCNAME,
			proto: beacondnspb.RRType_RR_TYPE_CNAME,
		},
		{
			name:  "MX record",
			model: model.RRTypeMX,
			proto: beacondnspb.RRType_RR_TYPE_MX,
		},
		{
			name:  "NS record",
			model: model.RRTypeNS,
			proto: beacondnspb.RRType_RR_TYPE_NS,
		},
		{
			name:  "PTR record",
			model: model.RRTypePTR,
			proto: beacondnspb.RRType_RR_TYPE_PTR,
		},
		{
			name:  "SOA record",
			model: model.RRTypeSOA,
			proto: beacondnspb.RRType_RR_TYPE_SOA,
		},
		{
			name:  "SRV record",
			model: model.RRTypeSRV,
			proto: beacondnspb.RRType_RR_TYPE_SRV,
		},
		{
			name:  "TXT record",
			model: model.RRTypeTXT,
			proto: beacondnspb.RRType_RR_TYPE_TXT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ToProto
			assert.Equal(t, tt.proto, RRTypeToProto(tt.model))

			// Test FromProto
			assert.Equal(t, tt.model, RRTypeFromProto(tt.proto))
		})
	}
}
