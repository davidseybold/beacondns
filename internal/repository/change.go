package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"

	"github.com/davidseybold/beacondns/internal/beaconerr"
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
	case model.ChangeTypeResponsePolicy:
		changeData = change.ResponsePolicyChange
	}

	changeDataJSON, err := json.Marshal(changeData)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to marshal change data", err)
	}

	row := r.db.QueryRow(ctx, insertChangeQuery, change.ID, change.Type, changeDataJSON, change.Status)

	err = row.Scan(&change.SubmittedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, beaconerr.ErrZoneAlreadyExists("zone already exists")
			}
		}

		return nil, beaconerr.ErrInternalError("failed to insert change into database", err)
	}

	return &change, nil
}

func (r *PostgresChangeRepository) GetChange(ctx context.Context, id uuid.UUID) (*model.Change, error) {
	row := r.db.QueryRow(ctx, getChangeQuery, id)

	var change model.Change
	var changeData []byte
	err := row.Scan(&change.ID, &change.Type, &changeData, &change.SubmittedAt)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, beaconerr.ErrNoSuchChange("change not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get change from database", err)
	}

	switch change.Type {
	case model.ChangeTypeZone:
		var zoneChange model.ZoneChange
		err = json.Unmarshal(changeData, &zoneChange)
		if err != nil {
			return nil, beaconerr.ErrInternalError("failed to unmarshal zone change data", err)
		}
		change.ZoneChange = &zoneChange
	case model.ChangeTypeResponsePolicy:
		var responsePolicyChange model.ResponsePolicyChange
		err = json.Unmarshal(changeData, &responsePolicyChange)
		if err != nil {
			return nil, beaconerr.ErrInternalError("failed to unmarshal response policy change data", err)
		}
		change.ResponsePolicyChange = &responsePolicyChange
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
		return beaconerr.ErrInternalError("failed to update change status", err)
	}
	return nil
}

func (r *PostgresChangeRepository) GetChangeToProcess(ctx context.Context) (*model.Change, error) {
	row := r.db.QueryRow(ctx, getChangeToProcessQuery)

	var change model.Change
	var changeData []byte
	err := row.Scan(&change.ID, &change.Type, &changeData, &change.SubmittedAt)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, beaconerr.ErrNoSuchChange("change not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to scan change row", err)
	}

	switch change.Type {
	case model.ChangeTypeZone:
		var zoneChange model.ZoneChange
		err = json.Unmarshal(changeData, &zoneChange)
		if err != nil {
			return nil, beaconerr.ErrInternalError("failed to unmarshal zone change data", err)
		}
		change.ZoneChange = &zoneChange
	case model.ChangeTypeResponsePolicy:
		var responsePolicyChange model.ResponsePolicyChange
		err = json.Unmarshal(changeData, &responsePolicyChange)
		if err != nil {
			return nil, beaconerr.ErrInternalError("failed to unmarshal response policy change data", err)
		}
		change.ResponsePolicyChange = &responsePolicyChange
	}

	return &change, nil
}
