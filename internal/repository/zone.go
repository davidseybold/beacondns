package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"

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

	insertChangeQuery = `
		INSERT INTO changes (id, zone_id, actions, status)
		VALUES ($1, $2, $3, $4)
		RETURNING submitted_at
	`

	getChangeQuery = `
		SELECT id, zone_id, actions, status, submitted_at
		FROM changes
		WHERE id = $1
	`

	getChangesByZoneQuery = `
		SELECT id, zone_id, actions, status, submitted_at
		FROM changes c
		INNER JOIN zones z ON z.id = c.zone_id
		WHERE z.name = $1
	`

	updateChangeStatusQuery = `
		UPDATE changes
		SET status = $2
		WHERE id = $1
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

	CreateChange(
		ctx context.Context,
		change model.Change,
	) (*model.Change, error)
	GetChange(ctx context.Context, id uuid.UUID) (*model.Change, error)
	GetChangesByZone(ctx context.Context, zoneName string) ([]model.Change, error)
	UpdateChangeStatus(ctx context.Context, id uuid.UUID, status model.ChangeStatus) error
}

type PostgresZoneRepository struct {
	db postgres.Queryer
}

var _ ZoneRepository = (*PostgresZoneRepository)(nil)

func (p *PostgresZoneRepository) CreateZone(ctx context.Context, zone *model.Zone) (*model.ZoneInfo, error) {
	if _, err := p.db.Exec(ctx, insertZoneQuery, zone.ID, zone.Name); err != nil {
		return nil, handleError(err, "failed to create zone: %w", err)
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
		return handleError(err, "failed to insert resource record set: %w", err)
	}

	for _, rr := range recordSet.ResourceRecords {
		_, err = p.db.Exec(ctx, insertResourceRecordQuery, id, rr.Value)
		if err != nil {
			return handleError(err, "failed to insert resource record: %w", err)
		}
	}

	return nil
}

func (p *PostgresZoneRepository) DeleteZone(ctx context.Context, name string) error {
	ct, err := p.db.Exec(ctx, deleteZoneQuery, name)
	if err != nil {
		return handleError(err, "failed to delete zone: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return ErrEntityNotFound
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
		return nil, fmt.Errorf("failed to upsert resource record set: %w", err)
	}

	_, err = p.db.Exec(ctx, deleteResourceRecordForSetQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing resource records: %w", err)
	}

	for _, rr := range recordSet.ResourceRecords {
		_, err = p.db.Exec(ctx, insertResourceRecordQuery, id, rr.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to insert resource record: %w", err)
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
		return fmt.Errorf("failed to delete resource record set: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return ErrEntityNotFound
	}

	return nil
}

func (p *PostgresZoneRepository) GetZone(ctx context.Context, name string) (*model.Zone, error) {
	row := p.db.QueryRow(ctx, selectZoneQuery, name)
	var zone model.Zone
	err := row.Scan(&zone.ID, &zone.Name)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEntityNotFound
	} else if err != nil {
		return nil, handleError(err, "failed to get zone: %w", err)
	}

	recordSets, err := p.GetZoneResourceRecordSets(ctx, zone.Name)
	if err != nil {
		return nil, handleError(err, "failed to get resource record sets for zone: %w", err)
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
		return nil, handleError(err, "failed to get resource record sets for zone: %w", err)
	}
	defer rows.Close()

	var recordSets []model.ResourceRecordSet
	for rows.Next() {
		var recordSet model.ResourceRecordSet
		var rrSetID uuid.UUID
		err = rows.Scan(&rrSetID, &recordSet.Name, &recordSet.Type, &recordSet.TTL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, handleError(err, "failed to scan resource record set: %w", err)
		} else if errors.Is(err, sql.ErrNoRows) {
			continue
		}

		var records []model.ResourceRecord
		records, err = p.getResourceRecordSetRecords(ctx, rrSetID)
		if err != nil {
			return nil, handleError(err, "failed to get resource records for resource record set: %w", err)
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
		return nil, ErrEntityNotFound
	} else if err != nil {
		return nil, handleError(err, "failed to get zone: %w", err)
	}

	return &zone, nil
}

func (p *PostgresZoneRepository) ListZoneInfos(ctx context.Context) ([]model.ZoneInfo, error) {
	rows, err := p.db.Query(ctx, selectZoneInfosQuery)
	if err != nil {
		return nil, handleError(err, "failed to list zone infos: %w", err)
	}
	defer rows.Close()

	var zoneInfos []model.ZoneInfo
	for rows.Next() {
		var zoneInfo model.ZoneInfo
		err = rows.Scan(&zoneInfo.ID, &zoneInfo.Name, &zoneInfo.ResourceRecordSetCount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, handleError(err, "failed to scan zone info: %w", err)
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
		return nil, ErrEntityNotFound
	} else if err != nil {
		return nil, handleError(err, "failed to get resource record set: %w", err)
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
		return nil, handleError(err, "failed to get resource records for resource record set: %w", err)
	}
	defer rows.Close()

	var records []model.ResourceRecord
	for rows.Next() {
		var rr model.ResourceRecord
		err = rows.Scan(&rr.Value)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, handleError(err, "failed to scan resource record: %w", err)
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
		return nil, handleError(err, "failed to get resource record sets by name: %w", err)
	}
	defer rows.Close()

	var recordSets []model.ResourceRecordSet
	for rows.Next() {
		var recordSet model.ResourceRecordSet
		err = rows.Scan(&recordSet.Name, &recordSet.Type, &recordSet.TTL)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, handleError(err, "failed to scan resource record set: %w", err)
		} else if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		recordSets = append(recordSets, recordSet)
	}

	return recordSets, nil
}

func (p *PostgresZoneRepository) CreateChange(
	ctx context.Context,
	change model.Change,
) (*model.Change, error) {
	actionsJSON, err := json.Marshal(change.Actions)
	if err != nil {
		return nil, handleError(err, "failed to marshal change data: %w", err)
	}

	row := p.db.QueryRow(ctx, insertChangeQuery, change.ID, change.ZoneID, actionsJSON, change.Status)

	err = row.Scan(&change.SubmittedAt)
	if err != nil {
		return nil, handleError(err, "failed to insert change into database: %w", err)
	}

	return &change, nil
}

func (p *PostgresZoneRepository) GetChange(ctx context.Context, id uuid.UUID) (*model.Change, error) {
	row := p.db.QueryRow(ctx, getChangeQuery, id)

	var change model.Change
	var actionsJSON []byte
	err := row.Scan(&change.ID, &change.ZoneID, &actionsJSON, &change.Status, &change.SubmittedAt)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEntityNotFound
	} else if err != nil {
		return nil, handleError(err, "failed to get change from database: %w", err)
	}

	err = json.Unmarshal(actionsJSON, &change.Actions)
	if err != nil {
		return nil, handleError(err, "failed to unmarshal change actions: %w", err)
	}

	return &change, nil
}

func (p *PostgresZoneRepository) UpdateChangeStatus(
	ctx context.Context,
	id uuid.UUID,
	status model.ChangeStatus,
) error {
	_, err := p.db.Exec(ctx, updateChangeStatusQuery, id, status)
	if err != nil {
		return handleError(err, "failed to update change status: %w", err)
	}
	return nil
}

func (p *PostgresZoneRepository) GetChangesByZone(ctx context.Context, zoneName string) ([]model.Change, error) {
	rows, err := p.db.Query(ctx, getChangesByZoneQuery, zoneName)
	if err != nil {
		return nil, handleError(err, "failed to get changes by zone: %w", err)
	}
	defer rows.Close()

	var changes []model.Change
	for rows.Next() {
		var change model.Change
		var actionsJSON []byte
		err = rows.Scan(&change.ID, &change.ZoneID, &actionsJSON, &change.Status, &change.SubmittedAt)
		if err != nil {
			return nil, handleError(err, "failed to scan change row: %w", err)
		}

		err = json.Unmarshal(actionsJSON, &change.Actions)
		if err != nil {
			return nil, handleError(err, "failed to unmarshal change actions: %w", err)
		}

		changes = append(changes, change)
	}

	return changes, nil
}
