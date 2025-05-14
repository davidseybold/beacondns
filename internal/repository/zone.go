package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/db/postgres"
	"github.com/davidseybold/beacondns/internal/model"
)

const (
	insertZoneQuery = "INSERT INTO zones(id, name) VALUES ($1, $2);"

	insertResourceRecordSetQuery = "INSERT INTO resource_record_sets (zone_id, name, record_type, ttl) VALUES ($1, $2, $3, $4) RETURNING id;"
	insertResourceRecordQuery    = "INSERT INTO resource_records (resource_record_set_id, value) VALUES ($1, $2);"
)

type ZoneRepository interface {
	CreateZone(ctx context.Context, params CreateZoneParams) error
}

type PostgresZoneRepository struct {
	db postgres.Queryer
}

type CreateZoneParams struct {
	Zone model.Zone
	SOA  model.ResourceRecordSet
	NS   model.ResourceRecordSet
}

func (p *PostgresZoneRepository) CreateZone(ctx context.Context, params CreateZoneParams) error {
	if _, err := p.db.Exec(ctx, insertZoneQuery, params.Zone.ID, params.Zone.Name); err != nil {
		return fmt.Errorf("failed to create zone %s: %w", params.Zone.Name, err)
	}

	if err := p.insertResourceRecordSets(ctx, params.Zone.ID, []model.ResourceRecordSet{params.SOA, params.NS}); err != nil {
		return fmt.Errorf("failed to create initial resource record sets for zone %s: %w", params.Zone.Name, err)
	}

	return nil
}

func (p *PostgresZoneRepository) insertResourceRecordSets(
	ctx context.Context,
	zoneID uuid.UUID,
	recordSets []model.ResourceRecordSet,
) error {
	var err error
	for _, recordSet := range recordSets {
		row := p.db.QueryRow(ctx, insertResourceRecordSetQuery, zoneID, recordSet.Name, recordSet.Type, recordSet.TTL)

		var id int
		err = row.Scan(&id)
		if err != nil {
			return fmt.Errorf(
				"failed to create resource record set %s of type %s for zone %s: %w",
				recordSet.Name,
				recordSet.Type,
				zoneID,
				err,
			)
		}

		for _, rr := range recordSet.ResourceRecords {
			_, err = p.db.Exec(ctx, insertResourceRecordQuery, id, rr.Value)
			if err != nil {
				return fmt.Errorf(
					"failed to create resource record with value %s for record set %d: %w",
					rr.Value,
					id,
					err,
				)
			}
		}
	}
	return nil
}
