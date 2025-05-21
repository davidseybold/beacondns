package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/davidseybold/beacondns/internal/db/postgres"
	"github.com/davidseybold/beacondns/internal/model"
)

const (
	insertZoneQuery = "INSERT INTO zones(id, name) VALUES ($1, $2);"

	insertResourceRecordSetQuery = "INSERT INTO resource_record_sets (id, zone_id, name, record_type, ttl) VALUES ($1, $2, $3, $4, $5);"
	insertResourceRecordQuery    = "INSERT INTO resource_records (resource_record_set_id, value) VALUES ($1, $2);"

	selectZoneInfoQuery = `
	SELECT z.id, z.name, count(rrs.*)
	FROM zones z
	INNER JOIN resource_record_sets rrs ON rrs.zone_id = z.id
	WHERE z.id = $1
	GROUP BY z.id
	`
)

type ZoneRepository interface {
	CreateZone(ctx context.Context, zone *model.Zone) error
	GetZoneInfo(ctx context.Context, id uuid.UUID) (*model.ZoneInfo, error)
}

type PostgresZoneRepository struct {
	db postgres.Queryer
}

func (p *PostgresZoneRepository) CreateZone(ctx context.Context, zone *model.Zone) error {
	if _, err := p.db.Exec(ctx, insertZoneQuery, zone.ID, zone.Name); err != nil {
		return fmt.Errorf("failed to create zone %s: %w", zone.Name, err)
	}

	if err := p.insertResourceRecordSets(ctx, zone.ID, zone.ResourceRecordSets); err != nil {
		return fmt.Errorf("failed to create initial resource record sets for zone %s: %w", zone.Name, err)
	}

	return nil
}

func (p *PostgresZoneRepository) insertResourceRecordSets(
	ctx context.Context,
	zoneID uuid.UUID,
	recordSets []model.ResourceRecordSet,
) error {
	for _, recordSet := range recordSets {
		_, err := p.db.Exec(
			ctx,
			insertResourceRecordSetQuery,
			recordSet.ID,
			zoneID,
			recordSet.Name,
			recordSet.Type,
			recordSet.TTL,
		)
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
			_, err = p.db.Exec(ctx, insertResourceRecordQuery, recordSet.ID, rr.Value)
			if err != nil {
				return fmt.Errorf(
					"failed to create resource record with value %s for record set %d: %w",
					rr.Value,
					recordSet.ID,
					err,
				)
			}
		}
	}
	return nil
}

func (p *PostgresZoneRepository) GetZoneInfo(ctx context.Context, id uuid.UUID) (*model.ZoneInfo, error) {
	row := p.db.QueryRow(ctx, selectZoneInfoQuery, id)
	var zone model.ZoneInfo
	err := row.Scan(&zone.ID, &zone.Name, &zone.ResourceRecordSetCount)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get zone %s: %w", id, err)
	}

	return &zone, nil
}
