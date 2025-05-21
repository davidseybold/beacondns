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

	insertResourceRecordSetQuery = "INSERT INTO resource_record_sets (id, zone_id, name, record_type, ttl) VALUES ($1, $2, $3, $4, $5) RETURNING id;"
	insertResourceRecordQuery    = "INSERT INTO resource_records (resource_record_set_id, value) VALUES ($1, $2);"

	selectZoneInfoQuery = `
	SELECT z.id, z.name, 
	       (SELECT COUNT(*) FROM resource_record_sets rrs WHERE rrs.zone_id = z.id) as record_count
	FROM zones z
	WHERE z.id = $1
	`

	selectZoneQuery = `
	SELECT z.id, z.name
	FROM zones z
	WHERE z.id = $1
	`

	selectZoneInfosQuery = `
	SELECT z.id, z.name, 
	       (SELECT COUNT(*) FROM resource_record_sets rrs WHERE rrs.zone_id = z.id) as record_count
	FROM zones z
	ORDER BY z.name
	LIMIT 1000
	`

	upsertResourceRecordSetQuery = `
	INSERT INTO resource_record_sets (id, zone_id, name, record_type, ttl)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (zone_id, name, record_type)
	DO UPDATE SET ttl = $5
	RETURNING id;
	`

	deleteResourceRecordSetQuery = `
	DELETE FROM resource_record_sets WHERE zone_id = $1 AND name = $2 AND record_type = $3;
	`

	deleteResourceRecordForSetQuery = `
	DELETE FROM resource_records WHERE resource_record_set_id = $1;
	`

	selectResourceRecordSetQuery = `
	SELECT rrs.id, rrs.name, rrs.record_type, rrs.ttl
	FROM resource_record_sets rrs
	WHERE rrs.id = $1
	`

	selectResourceRecordsQuery = `
	SELECT rr.value
	FROM resource_records rr
	WHERE rr.resource_record_set_id = $1
	ORDER BY rr.value
	`

	selectResourceRecordSetsByNameQuery = `
	SELECT rrs.id, rrs.name, rrs.record_type, rrs.ttl
	FROM resource_record_sets rrs
	WHERE rrs.zone_id = $1 AND rrs.name = $2
	ORDER BY rrs.record_type
	`

	selectResourceRecordSetsForZoneQuery = `
	SELECT rrs.id, rrs.name, rrs.record_type, rrs.ttl
	FROM resource_record_sets rrs
	WHERE rrs.zone_id = $1
	ORDER BY rrs.name, rrs.record_type
	`
)

type ZoneRepository interface {
	CreateZone(ctx context.Context, zone *model.Zone) error
	GetZone(ctx context.Context, id uuid.UUID) (*model.Zone, error)
	GetZoneInfo(ctx context.Context, id uuid.UUID) (*model.ZoneInfo, error)
	ListZoneInfos(ctx context.Context) ([]model.ZoneInfo, error)
	GetResourceRecordSet(ctx context.Context, id uuid.UUID) (*model.ResourceRecordSet, error)
	GetZoneResourceRecordSets(ctx context.Context, zoneID uuid.UUID) ([]model.ResourceRecordSet, error)
	InsertResourceRecordSet(ctx context.Context, zoneID uuid.UUID, recordSet *model.ResourceRecordSet) error
	UpsertResourceRecordSet(ctx context.Context, zoneID uuid.UUID, recordSet *model.ResourceRecordSet) error
	DeleteResourceRecordSet(ctx context.Context, zoneID uuid.UUID, recordSet *model.ResourceRecordSet) error
}

type PostgresZoneRepository struct {
	db postgres.Queryer
}

var _ ZoneRepository = (*PostgresZoneRepository)(nil)

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
		err := p.InsertResourceRecordSet(ctx, zoneID, &recordSet)
		if err != nil {
			return fmt.Errorf("failed to insert resource record set: %w", err)
		}
	}

	return nil
}

func (p *PostgresZoneRepository) InsertResourceRecordSet(
	ctx context.Context,
	zoneID uuid.UUID,
	recordSet *model.ResourceRecordSet,
) error {
	row := p.db.QueryRow(
		ctx,
		insertResourceRecordSetQuery,
		recordSet.ID,
		zoneID,
		recordSet.Name,
		recordSet.Type,
		recordSet.TTL,
	)

	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert resource record set: %w", err)
	}

	for _, rr := range recordSet.ResourceRecords {
		_, err = p.db.Exec(ctx, insertResourceRecordQuery, id, rr.Value)
		if err != nil {
			return fmt.Errorf("failed to insert resource record: %w", err)
		}
	}

	return nil
}

func (p *PostgresZoneRepository) UpsertResourceRecordSet(
	ctx context.Context,
	zoneID uuid.UUID,
	recordSet *model.ResourceRecordSet,
) error {
	row := p.db.QueryRow(
		ctx,
		upsertResourceRecordSetQuery,
		recordSet.ID,
		zoneID,
		recordSet.Name,
		recordSet.Type,
		recordSet.TTL,
	)

	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to upsert resource record set: %w", err)
	}

	_, err = p.db.Exec(ctx, deleteResourceRecordForSetQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete existing resource records: %w", err)
	}

	for _, rr := range recordSet.ResourceRecords {
		_, err = p.db.Exec(ctx, insertResourceRecordQuery, id, rr.Value)
		if err != nil {
			return fmt.Errorf("failed to insert resource record: %w", err)
		}
	}

	return nil
}

func (p *PostgresZoneRepository) DeleteResourceRecordSet(
	ctx context.Context,
	zoneID uuid.UUID,
	recordSet *model.ResourceRecordSet,
) error {
	_, err := p.db.Exec(ctx, deleteResourceRecordSetQuery, zoneID, recordSet.Name, recordSet.Type)
	if err != nil {
		return fmt.Errorf("failed to delete resource record set: %w", err)
	}

	return nil
}

func (p *PostgresZoneRepository) GetZone(ctx context.Context, id uuid.UUID) (*model.Zone, error) {
	row := p.db.QueryRow(ctx, selectZoneQuery, id)
	var zone model.Zone
	err := row.Scan(&zone.ID, &zone.Name)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get zone %s: %w", id, err)
	}

	recordSets, err := p.GetZoneResourceRecordSets(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource record sets for zone %s: %w", id, err)
	}
	zone.ResourceRecordSets = recordSets

	return &zone, nil
}

func (p *PostgresZoneRepository) GetZoneResourceRecordSets(
	ctx context.Context,
	zoneID uuid.UUID,
) ([]model.ResourceRecordSet, error) {
	rows, err := p.db.Query(ctx, selectResourceRecordSetsForZoneQuery, zoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource record sets for zone %s: %w", zoneID, err)
	}
	defer rows.Close()

	var recordSets []model.ResourceRecordSet
	for rows.Next() {
		var recordSet model.ResourceRecordSet
		err = rows.Scan(&recordSet.ID, &recordSet.Name, &recordSet.Type, &recordSet.TTL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to scan resource record set: %w", err)
		} else if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		recordSets = append(recordSets, recordSet)
	}

	for i, recordSet := range recordSets {
		var records []model.ResourceRecord
		records, err = p.getResourceRecordSetRecords(ctx, recordSet.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource records for resource record set %s: %w", recordSet.ID, err)
		}
		recordSets[i].ResourceRecords = records
	}

	return recordSets, nil
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

func (p *PostgresZoneRepository) ListZoneInfos(ctx context.Context) ([]model.ZoneInfo, error) {
	rows, err := p.db.Query(ctx, selectZoneInfosQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list zone infos: %w", err)
	}
	defer rows.Close()

	var zoneInfos []model.ZoneInfo
	for rows.Next() {
		var zoneInfo model.ZoneInfo
		err = rows.Scan(&zoneInfo.ID, &zoneInfo.Name, &zoneInfo.ResourceRecordSetCount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to scan zone info: %w", err)
		} else if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		zoneInfos = append(zoneInfos, zoneInfo)
	}

	return zoneInfos, nil
}

func (p *PostgresZoneRepository) GetResourceRecordSet(
	ctx context.Context,
	id uuid.UUID,
) (*model.ResourceRecordSet, error) {
	row := p.db.QueryRow(ctx, selectResourceRecordSetQuery, id)
	var recordSet model.ResourceRecordSet
	err := row.Scan(&recordSet.ID, &recordSet.Name, &recordSet.Type, &recordSet.TTL)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get resource record set %s: %w", id, err)
	}

	records, err := p.getResourceRecordSetRecords(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource records for resource record set %s: %w", id, err)
	}
	recordSet.ResourceRecords = records

	return &recordSet, nil
}

func (p *PostgresZoneRepository) getResourceRecordSetRecords(
	ctx context.Context,
	id uuid.UUID,
) ([]model.ResourceRecord, error) {
	rows, err := p.db.Query(ctx, selectResourceRecordsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource records for resource record set %s: %w", id, err)
	}
	defer rows.Close()

	var records []model.ResourceRecord
	for rows.Next() {
		var rr model.ResourceRecord
		err = rows.Scan(&rr.Value)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to scan resource record: %w", err)
		}
		records = append(records, rr)
	}

	return records, nil
}

func (p *PostgresZoneRepository) GetResourceRecordSetsByName(
	ctx context.Context,
	zoneID uuid.UUID,
	name string,
) ([]model.ResourceRecordSet, error) {
	rows, err := p.db.Query(ctx, selectResourceRecordSetsByNameQuery, zoneID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource record sets by name %s: %w", name, err)
	}
	defer rows.Close()

	var recordSets []model.ResourceRecordSet
	for rows.Next() {
		var recordSet model.ResourceRecordSet
		err = rows.Scan(&recordSet.ID, &recordSet.Name, &recordSet.Type, &recordSet.TTL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to scan resource record set: %w", err)
		} else if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		recordSets = append(recordSets, recordSet)
	}

	return recordSets, nil
}
