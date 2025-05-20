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
	insertChangeQuery = `
		INSERT INTO changes (id, type, data, status)
		VALUES ($1, $2, $3, $4)
		RETURNING submitted_at
	`

	getChangeQuery = `
		SELECT id, type, data, submitted_at
		FROM changes
		WHERE id = $1
	`

	updateChangeStatusQuery = `
		UPDATE changes
		SET status = $2
		WHERE id = $1
	`

	getChangeToProcessQuery = `
		UPDATE changes
		SET lock_expires = NOW() + INTERVAL '5 minutes'
		WHERE id = (
			SELECT id
			FROM changes
			WHERE status = 'PENDING'
    		AND (lock_expires < NOW() OR lock_expires IS NULL)
			ORDER BY submitted_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, type, data, submitted_at
	`
)

type ChangeRepository interface {
	CreateChange(
		ctx context.Context,
		change model.Change,
	) (*model.Change, error)
	GetChange(ctx context.Context, id uuid.UUID) (*model.Change, error)
	UpdateChangeStatus(ctx context.Context, id uuid.UUID, status model.ChangeStatus) error
	GetChangeToProcess(ctx context.Context) (*model.Change, error)
}

type PostgresChangeRepository struct {
	db postgres.Queryer
}

var _ ChangeRepository = (*PostgresChangeRepository)(nil)

func (r *PostgresChangeRepository) CreateChange(
	ctx context.Context,
	change model.Change,
) (*model.Change, error) {
	var changeData any
	switch change.Type {
	case model.ChangeTypeZone:
		changeData = change.ZoneChange
	default:
		return nil, fmt.Errorf("failed to create change: unknown change type: %s", change.Type)
	}

	changeDataJSON, err := json.Marshal(changeData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal change data for change %s: %w", change.ID, err)
	}

	row := r.db.QueryRow(ctx, insertChangeQuery, change.ID, change.Type, changeDataJSON, change.Status)

	err = row.Scan(&change.SubmittedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert change %s into database: %w", change.ID, err)
	}

	return &change, nil
}

func (r *PostgresChangeRepository) GetChange(ctx context.Context, id uuid.UUID) (*model.Change, error) {
	row := r.db.QueryRow(ctx, getChangeQuery, id)

	var change model.Change
	var changeData []byte
	err := row.Scan(&change.ID, &change.Type, &changeData, &change.SubmittedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get change %s from database: %w", id, err)
	}

	switch change.Type {
	case model.ChangeTypeZone:
		var zoneChange model.ZoneChange
		err = json.Unmarshal(changeData, &zoneChange)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal zone change data for change %s: %w", id, err)
		}
		change.ZoneChange = &zoneChange
	default:
		return nil, fmt.Errorf("failed to process change %s: unknown change type: %s", id, change.Type)
	}

	return &change, nil
}

func (r *PostgresChangeRepository) UpdateChangeStatus(
	ctx context.Context,
	id uuid.UUID,
	status model.ChangeStatus,
) error {
	_, err := r.db.Exec(ctx, updateChangeStatusQuery, id, status)
	if err != nil {
		return fmt.Errorf("failed to update change status for change %s: %w", id, err)
	}
	return nil
}

func (r *PostgresChangeRepository) GetChangeToProcess(ctx context.Context) (*model.Change, error) {
	row := r.db.QueryRow(ctx, getChangeToProcessQuery)

	var change model.Change
	var changeData []byte
	err := row.Scan(&change.ID, &change.Type, &changeData, &change.SubmittedAt)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to scan change row: %w", err)
	}

	switch change.Type {
	case model.ChangeTypeZone:
		var zoneChange model.ZoneChange
		err = json.Unmarshal(changeData, &zoneChange)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal zone change data for change %s: %w", change.ID, err)
		}
		change.ZoneChange = &zoneChange
	default:
		return nil, fmt.Errorf("failed to process change %s: unknown change type: %s", change.ID, change.Type)
	}

	return &change, nil
}
