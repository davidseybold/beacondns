package repository

import (
	"context"
	"encoding/json"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
	"github.com/google/uuid"
)

const (
	insertDelegationSetQuery            = "INSERT INTO delegation_sets (id) VALUES ($1) RETURNING id;"
	insertDelegationSetNameServersQuery = "INSERT INTO delegation_set_nameservers (delegation_set_id, nameserver_id) VALUES($1, $2);"

	insertZoneQuery = "INSERT INTO zones(id, name, delegation_set_id, is_private) VALUES ($1, $2, $3, $4);"

	insertResourceRecordSetQuery = "INSERT INTO resource_record_sets (id, zone_id, name, record_type, ttl) VALUES ($1, $2, $3, $4, $5);"
	insertResourceRecordQuery    = "INSERT INTO resource_records (resource_record_set_id, value) VALUES ($1, $2);"

	// TODO: Update created_at to submitted_at
	insertZoneChangeQuery = `
	INSERT INTO zone_changes (id, zone_id, action)
	VALUES ($1, $2, $3) RETURNING created_at
	`
	insertResourceRecordSetChangeQuery = `
	INSERT INTO resource_record_set_changes (
    	zone_change_id, action, name, record_type, ttl, record_values, ordering
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	insertZoneChangeSyncQuery = `
	INSERT INTO zone_change_syncs (zone_change_id, nameserver_id, status)
	VALUES ($1, $2, $3)
	`
)

type ZoneRepository interface {
	CreateZone(ctx context.Context, params CreateZoneParams) (*domain.ChangeInfo, error)
}

type PostgresZoneRepository struct {
	db postgres.Queryer
}

type CreateZoneParams struct {
	Zone          domain.Zone
	DelegationSet *domain.DelegationSet
	SOA           domain.ResourceRecordSet
	NS            domain.ResourceRecordSet
	Change        domain.ZoneChange
	Syncs         []domain.ZoneChangeSync
}

func (p *PostgresZoneRepository) CreateZone(ctx context.Context, params CreateZoneParams) (*domain.ChangeInfo, error) {
	var delegationSetID *uuid.UUID
	if params.DelegationSet != nil {
		if err := p.insertDelegationSet(ctx, *params.DelegationSet); err != nil {
			return nil, err
		}
		delegationSetID = &params.DelegationSet.ID
	}

	if _, err := p.db.Exec(ctx, insertZoneQuery, params.Zone.ID, params.Zone.Name, delegationSetID, params.Zone.IsPrivate); err != nil {
		return nil, err
	}

	if err := p.insertResourceRecordSets(ctx, params.Zone.ID, []domain.ResourceRecordSet{params.SOA, params.NS}); err != nil {
		return nil, err
	}

	changeInfo, err := p.insertZoneChange(ctx, params.Change)
	if err != nil {
		return nil, err
	}

	for _, sync := range params.Syncs {
		if _, err := p.db.Exec(ctx, insertZoneChangeSyncQuery, sync.ZoneChangeID, sync.NameServerID, sync.Status); err != nil {
			return nil, err
		}
	}

	return changeInfo, nil
}

func (p *PostgresZoneRepository) insertDelegationSet(ctx context.Context, ds domain.DelegationSet) error {
	if _, err := p.db.Exec(ctx, insertDelegationSetQuery, ds.ID); err != nil {
		return err
	}

	for _, ns := range ds.NameServers {
		if _, err := p.db.Exec(ctx, insertDelegationSetNameServersQuery, ds.ID, ns.ID); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresZoneRepository) insertResourceRecordSets(ctx context.Context, zoneID uuid.UUID, recordSets []domain.ResourceRecordSet) error {
	for _, recordSet := range recordSets {

		if _, err := p.db.Exec(ctx, insertResourceRecordSetQuery, recordSet.ID, zoneID, recordSet.Name, recordSet.Type, recordSet.TTL); err != nil {
			return err
		}

		for _, rr := range recordSet.ResourceRecords {
			if _, err := p.db.Exec(ctx, insertResourceRecordQuery, recordSet.ID, rr.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *PostgresZoneRepository) insertZoneChange(ctx context.Context, change domain.ZoneChange) (*domain.ChangeInfo, error) {
	row := p.db.QueryRow(ctx, insertZoneChangeQuery, change.ID, change.ZoneID, change.Action)

	changeInfo := &domain.ChangeInfo{
		ID:     change.ID,
		Status: domain.ChangeSyncStatusPending,
	}
	if err := row.Scan(&changeInfo.SubmittedAt); err != nil {
		return nil, err
	}

	for i, rrcChange := range change.Changes {
		values := getRecordSetValues(rrcChange.ResourceRecordSet)
		b, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}

		if _, err := p.db.Exec(
			ctx,
			insertResourceRecordSetChangeQuery,
			change.ID,
			rrcChange.Action,
			rrcChange.ResourceRecordSet.Name,
			rrcChange.ResourceRecordSet.Type,
			rrcChange.ResourceRecordSet.TTL,
			b, // record values
			i+1,
		); err != nil {
			return nil, err
		}
	}

	return changeInfo, nil
}

func getRecordSetValues(rrs domain.ResourceRecordSet) []string {
	values := make([]string, len(rrs.ResourceRecords))
	for i, rr := range rrs.ResourceRecords {
		values[i] = rr.Value
	}
	return values
}
