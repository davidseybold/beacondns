package zone

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/davidseybold/beacondns/internal/model"
)

func TestDomainNameRule(t *testing.T) {
	tests := []struct {
		name    string
		zone    *model.Zone
		changes []model.ChangeAction
		wantErr bool
		err     error
	}{
		{
			name: "valid apex record",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "example.com",
						Type: model.RRTypeSOA,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid subdomain record",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "test.example.com",
						Type: model.RRTypeA,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid subdomain record",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "test.otherdomain.com",
						Type: model.RRTypeA,
					},
				},
			},
			wantErr: true,
			err:     ErrOutsideZone,
		},
		{
			name: "valid wildcard at leftmost label",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "*.example.com",
						Type: model.RRTypeA,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid wildcard at leftmost label with trailing dot",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "*.example.com.",
						Type: model.RRTypeA,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid wildcard in middle of name",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "sub.*.example.com",
						Type: model.RRTypeA,
					},
				},
			},
			wantErr: true,
			err:     ErrInvalidWildcard,
		},
		{
			name: "invalid wildcard at end of name",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "sub.example.*",
						Type: model.RRTypeA,
					},
				},
			},
			wantErr: true,
			err:     ErrInvalidWildcard,
		},
		{
			name: "invalid wildcard in middle of label",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "sub*.example.com",
						Type: model.RRTypeA,
					},
				},
			},
			wantErr: true,
			err:     ErrInvalidWildcard,
		},
		{
			name: "record outside zone",
			zone: &model.Zone{
				Name: "example.com",
			},
			changes: []model.ChangeAction{
				{
					ResourceRecordSet: &model.ResourceRecordSet{
						Name: "test.otherdomain.com",
						Type: model.RRTypeA,
					},
				},
			},
			wantErr: true,
			err:     ErrOutsideZone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domainNameRule(tt.zone, tt.changes)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorAs(t, err, &tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
