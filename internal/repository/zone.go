package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/db/postgres"
	"github.com/davidseybold/beacondns/internal/model"
)

const (
	insertZoneQuery              = "INSERT INTO zones(id, name) VALUES ($1, $2);"
	deleteZoneQuery              = "DELETE FROM zones WHERE name = $1;"
	insertResourceRecordSetQuery = "INSERT INTO resource_record_sets (id, zone_id, name, record_type, ttl) VALUES ($1, $2, $3, $4, $5) RETURNING id;"
	insertResourceRecordQuery    = "INSERT INTO resource_records (resource_record_set_id, value) VALUES ($1, $2);"

	selectZoneInfoQuery = `
	SELECT z.id, z.name, 
	       (SELECT COUNT(*) FROM resource_record_sets rrs WHERE rrs.zone_id = z.id) as record_count
	FROM zones z
	WHERE z.name = $1
	`

	selectZoneQuery = `
	SELECT z.id, z.name
	FROM zones z
	WHERE z.name = $1
	`

	selectZoneInfosQuery = `
	SELECT z.id, z.name, 
	       (SELECT COUNT(*) FROM resource_record_sets rrs WHERE rrs.zone_id = z.id) as record_count
	FROM zones z
	ORDER BY z.name
	LIMIT 1000
	`

	upsertResourceRecordSetQuery = `
	WITH zone_lookup AS (
		SELECT id FROM zones WHERE name = $2
	)
	INSERT INTO resource_record_sets (id, zone_id, name, record_type, ttl)
	SELECT $1, zone_lookup.id, $3, $4, $5
	FROM zone_lookup
	ON CONFLICT (zone_id, name, record_type)
	DO UPDATE SET ttl = $5
	RETURNING id;
	`

	deleteResourceRecordSetQuery = `
	DELETE FROM resource_record_sets rrs
	USING zones z
	WHERE rrs.zone_id = z.id
		AND z.name = $1
		AND rrs.name = $2
		AND rrs.record_type = $3;
	`

	deleteResourceRecordForSetQuery = `
	DELETE FROM resource_records WHERE resource_record_set_id = $1;
	`

	selectResourceRecordSetQuery = `
	SELECT rrs.id, rrs.name, rrs.record_type, rrs.ttl
	FROM resource_record_sets rrs
	INNER JOIN zones z ON z.id = rrs.zone_id
	WHERE z.name = $1 AND rrs.name = $2 AND rrs.record_type = $3
	`

	selectResourceRecordsQuery = `
	SELECT rr.value
	FROM resource_records rr
	WHERE rr.resource_record_set_id = $1
	ORDER BY rr.value
	`

	selectResourceRecordSetsByNameQuery = `
	SELECT rrs.name, rrs.record_type, rrs.ttl
	FROM resource_record_sets rrs
	INNER JOIN zones z ON z.id = rrs.zone_id
	WHERE z.name = $1 AND rrs.name = $2
	ORDER BY rrs.record_type
	`

	selectResourceRecordSetsForZoneQuery = `
	SELECT rrs.id, rrs.name, rrs.record_type, rrs.ttl
	FROM resource_record_sets rrs
	INNER JOIN zones z ON z.id = rrs.zone_id
	WHERE z.name = $1
	ORDER BY rrs.name, rrs.record_type
	`
)

type ZoneRepository interface {
	CreateZone(ctx context.Context, zone *model.Zone) (*model.ZoneInfo, error)
	DeleteZone(ctx context.Context, name string) error
	GetZone(ctx context.Context, name string) (*model.Zone, error)
	GetZoneInfo(ctx context.Context, name string) (*model.ZoneInfo, error)
	ListZoneInfos(ctx context.Context) ([]model.ZoneInfo, error)
	GetResourceRecordSet(
		ctx context.Context,
		zoneName string,
		name string,
		rrType model.RRType,
	) (*model.ResourceRecordSet, error)
	GetZoneResourceRecordSets(ctx context.Context, zoneName string) ([]model.ResourceRecordSet, error)
	UpsertResourceRecordSet(
		ctx context.Context,
		zoneName string,
		recordSet *model.ResourceRecordSet,
	) (*model.ResourceRecordSet, error)
	DeleteResourceRecordSet(ctx context.Context, zoneName string, name string, rrType model.RRType) error
}

type PostgresZoneRepository struct {
	db postgres.Queryer
}

var _ ZoneRepository = (*PostgresZoneRepository)(nil)

func (p *PostgresZoneRepository) CreateZone(ctx context.Context, zone *model.Zone) (*model.ZoneInfo, error) {
	if _, err := p.db.Exec(ctx, insertZoneQuery, zone.ID, zone.Name); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, beaconerr.ErrZoneAlreadyExists("zone already exists")
			}
		}

		return nil, beaconerr.ErrInternalError("failed to create zone", err)
	}

	if err := p.insertResourceRecordSets(ctx, zone.ID, zone.ResourceRecordSets); err != nil {
		return nil, err
	}

	return &model.ZoneInfo{
		ID:                     zone.ID,
		Name:                   zone.Name,
		ResourceRecordSetCount: len(zone.ResourceRecordSets),
	}, nil
}

func (p *PostgresZoneRepository) insertResourceRecordSets(
	ctx context.Context,
	zoneID uuid.UUID,
	recordSets []model.ResourceRecordSet,
) error {
	for _, recordSet := range recordSets {
		err := p.insertResourceRecordSet(ctx, zoneID, &recordSet)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresZoneRepository) insertResourceRecordSet(
	ctx context.Context,
	zoneID uuid.UUID,
	recordSet *model.ResourceRecordSet,
) error {
	row := p.db.QueryRow(
		ctx,
		insertResourceRecordSetQuery,
		uuid.New(),
		zoneID,
		recordSet.Name,
		recordSet.Type,
		recordSet.TTL,
	)

	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		return beaconerr.ErrInternalError("failed to insert resource record set", err)
	}

	for _, rr := range recordSet.ResourceRecords {
		_, err = p.db.Exec(ctx, insertResourceRecordQuery, id, rr.Value)
		if err != nil {
			return beaconerr.ErrInternalError("failed to insert resource record", err)
		}
	}

	return nil
}

func (p *PostgresZoneRepository) DeleteZone(ctx context.Context, name string) error {
	ct, err := p.db.Exec(ctx, deleteZoneQuery, name)
	if err != nil {
		return beaconerr.ErrInternalError("failed to delete zone", err)
	}

	if ct.RowsAffected() == 0 {
		return beaconerr.ErrNoSuchZone("zone not found")
	}

	return nil
}

func (p *PostgresZoneRepository) UpsertResourceRecordSet(
	ctx context.Context,
	zoneName string,
	recordSet *model.ResourceRecordSet,
) (*model.ResourceRecordSet, error) {
	row := p.db.QueryRow(
		ctx,
		upsertResourceRecordSetQuery,
		uuid.New(),
		zoneName,
		recordSet.Name,
		recordSet.Type,
		recordSet.TTL,
	)

	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to upsert resource record set", err)
	}

	_, err = p.db.Exec(ctx, deleteResourceRecordForSetQuery, id)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to delete existing resource records", err)
	}

	for _, rr := range recordSet.ResourceRecords {
		_, err = p.db.Exec(ctx, insertResourceRecordQuery, id, rr.Value)
		if err != nil {
			return nil, beaconerr.ErrInternalError("failed to insert resource record", err)
		}
	}

	return &model.ResourceRecordSet{
		Name:            recordSet.Name,
		Type:            recordSet.Type,
		TTL:             recordSet.TTL,
		ResourceRecords: recordSet.ResourceRecords,
	}, nil
}

func (p *PostgresZoneRepository) DeleteResourceRecordSet(
	ctx context.Context,
	zoneName string,
	name string,
	rrType model.RRType,
) error {
	ct, err := p.db.Exec(ctx, deleteResourceRecordSetQuery, zoneName, name, rrType)
	if err != nil {
		return beaconerr.ErrInternalError("failed to delete resource record set", err)
	}

	if ct.RowsAffected() == 0 {
		return beaconerr.ErrNoSuchResourceRecordSet("resource record set not found")
	}

	return nil
}

func (p *PostgresZoneRepository) GetZone(ctx context.Context, name string) (*model.Zone, error) {
	row := p.db.QueryRow(ctx, selectZoneQuery, name)
	var zone model.Zone
	err := row.Scan(&zone.ID, &zone.Name)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, beaconerr.ErrNoSuchZone("zone not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get zone", err)
	}

	recordSets, err := p.GetZoneResourceRecordSets(ctx, zone.Name)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get resource record sets for zone", err)
	}
	zone.ResourceRecordSets = recordSets

	return &zone, nil
}

func (p *PostgresZoneRepository) GetZoneResourceRecordSets(
	ctx context.Context,
	zoneName string,
) ([]model.ResourceRecordSet, error) {
	rows, err := p.db.Query(ctx, selectResourceRecordSetsForZoneQuery, zoneName)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get resource record sets for zone", err)
	}
	defer rows.Close()

	var recordSets []model.ResourceRecordSet
	for rows.Next() {
		var recordSet model.ResourceRecordSet
		var rrSetID uuid.UUID
		err = rows.Scan(&rrSetID, &recordSet.Name, &recordSet.Type, &recordSet.TTL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, beaconerr.ErrInternalError("failed to scan resource record set", err)
		} else if errors.Is(err, sql.ErrNoRows) {
			continue
		}

		var records []model.ResourceRecord
		records, err = p.getResourceRecordSetRecords(ctx, rrSetID)
		if err != nil {
			return nil, beaconerr.ErrInternalError("failed to get resource records for resource record set", err)
		}
		recordSet.ResourceRecords = records

		recordSets = append(recordSets, recordSet)
	}

	return recordSets, nil
}

func (p *PostgresZoneRepository) GetZoneInfo(ctx context.Context, name string) (*model.ZoneInfo, error) {
	row := p.db.QueryRow(ctx, selectZoneInfoQuery, name)
	var zone model.ZoneInfo
	err := row.Scan(&zone.ID, &zone.Name, &zone.ResourceRecordSetCount)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, beaconerr.ErrNoSuchZone("zone not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get zone", err)
	}

	return &zone, nil
}

func (p *PostgresZoneRepository) ListZoneInfos(ctx context.Context) ([]model.ZoneInfo, error) {
	rows, err := p.db.Query(ctx, selectZoneInfosQuery)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to list zone infos", err)
	}
	defer rows.Close()

	var zoneInfos []model.ZoneInfo
	for rows.Next() {
		var zoneInfo model.ZoneInfo
		err = rows.Scan(&zoneInfo.ID, &zoneInfo.Name, &zoneInfo.ResourceRecordSetCount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, beaconerr.ErrInternalError("failed to scan zone info", err)
		} else if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		zoneInfos = append(zoneInfos, zoneInfo)
	}

	return zoneInfos, nil
}

func (p *PostgresZoneRepository) GetResourceRecordSet(
	ctx context.Context,
	zoneName string,
	name string,
	rrType model.RRType,
) (*model.ResourceRecordSet, error) {
	row := p.db.QueryRow(ctx, selectResourceRecordSetQuery, zoneName, name, rrType)
	var recordSet model.ResourceRecordSet
	var rrSetID uuid.UUID
	err := row.Scan(&rrSetID, &recordSet.Name, &recordSet.Type, &recordSet.TTL)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, beaconerr.ErrNoSuchResourceRecordSet("resource record set not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get resource record set", err)
	}

	records, err := p.getResourceRecordSetRecords(ctx, rrSetID)
	if err != nil {
		return nil, err
	}
	recordSet.ResourceRecords = records

	return &recordSet, nil
}

func (p *PostgresZoneRepository) getResourceRecordSetRecords(
	ctx context.Context,
	rrSetID uuid.UUID,
) ([]model.ResourceRecord, error) {
	rows, err := p.db.Query(ctx, selectResourceRecordsQuery, rrSetID)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get resource records for resource record set", err)
	}
	defer rows.Close()

	var records []model.ResourceRecord
	for rows.Next() {
		var rr model.ResourceRecord
		err = rows.Scan(&rr.Value)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, beaconerr.ErrInternalError("failed to scan resource record", err)
		}
		records = append(records, rr)
	}

	return records, nil
}

func (p *PostgresZoneRepository) GetResourceRecordSetsByName(
	ctx context.Context,
	zoneName string,
	name string,
) ([]model.ResourceRecordSet, error) {
	rows, err := p.db.Query(ctx, selectResourceRecordSetsByNameQuery, zoneName, name)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get resource record sets by name", err)
	}
	defer rows.Close()

	var recordSets []model.ResourceRecordSet
	for rows.Next() {
		var recordSet model.ResourceRecordSet
		err = rows.Scan(&recordSet.Name, &recordSet.Type, &recordSet.TTL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, beaconerr.ErrInternalError("failed to scan resource record set", err)
		} else if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		recordSets = append(recordSets, recordSet)
	}

	return recordSets, nil
}
