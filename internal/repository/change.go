package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/db/postgres"
	"github.com/davidseybold/beacondns/internal/model"
)

const (
	insertChangeQuery = `
		INSERT INTO changes (id, type, data)
		VALUES ($1, $2, $3)
		RETURNING submitted_at
	`

	getChangeQuery = `
		SELECT id, type, data, submitted_at
		FROM changes
		WHERE id = $1
	`

	insertChangeTargetQuery = `
		INSERT INTO change_targets (change_id, server_id)
		VALUES ($1, $2)
	`

	getChangesWithPendingTargetsQuery = `
		SELECT DISTINCT c.id, c.type, c.data, c.submitted_at
		FROM changes c
		JOIN change_targets ct ON c.id = ct.change_id
		WHERE ct.status = 'PENDING'
	`

	getPendingTargetsForChangeQuery = `
		SELECT ct.status, ct.synced_at, s.id, s.type, s.hostname
		FROM change_targets ct
		INNER JOIN servers s ON ct.server_id = s.id
		INNER JOIN changes c ON ct.change_id = c.id
		WHERE change_id = $1 AND status = 'PENDING' ORDER BY c.submitted_at ASC
	`

	updateChangeTargetStatusQuery = `
		UPDATE change_targets ct
		SET status = $3, synced_at = CASE WHEN $3 = 'INSYNC' THEN CURRENT_TIMESTAMP ELSE synced_at END
		FROM servers s
		WHERE ct.change_id = $1 AND ct.server_id = s.id AND s.hostname = $2
	`
)

type ChangeRepository interface {
	CreateChange(
		ctx context.Context,
		change model.Change,
	) (*model.Change, error)
	GetChange(ctx context.Context, id uuid.UUID) (*model.Change, error)
	GetChangesWithPendingTargets(ctx context.Context) ([]*model.Change, error)
	GetPendingTargetsForChange(ctx context.Context, changeID uuid.UUID) ([]model.ChangeTarget, error)
	UpdateChangeTargetStatus(
		ctx context.Context,
		changeID uuid.UUID,
		hostName string,
		status model.ChangeTargetStatus,
	) error
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

	row := r.db.QueryRow(ctx, insertChangeQuery, change.ID, change.Type, changeDataJSON)

	err = row.Scan(&change.SubmittedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert change %s into database: %w", change.ID, err)
	}

	err = r.createChangeTargets(ctx, change.ID, change.Targets)
	if err != nil {
		return nil, fmt.Errorf("failed to create change targets for change %s: %w", change.ID, err)
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

func (r *PostgresChangeRepository) createChangeTargets(
	ctx context.Context,
	changeID uuid.UUID,
	targets []model.ChangeTarget,
) error {
	for _, target := range targets {
		_, err := r.db.Exec(ctx, insertChangeTargetQuery, changeID, target.Server.ID)
		if err != nil {
			return fmt.Errorf(
				"failed to create change target for change %s and hostname %s: %w",
				changeID,
				target.Server.HostName,
				err,
			)
		}
	}
	return nil
}

func (r *PostgresChangeRepository) GetChangesWithPendingTargets(ctx context.Context) ([]*model.Change, error) {
	var err error
	rows, err := r.db.Query(ctx, getChangesWithPendingTargetsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query changes with pending targets: %w", err)
	}
	defer rows.Close()

	var changes []*model.Change
	for rows.Next() {
		var change model.Change
		var changeData []byte
		err = rows.Scan(&change.ID, &change.Type, &changeData, &change.SubmittedAt)
		if err != nil {
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

		changes = append(changes, &change)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating change rows: %w", err)
	}

	return changes, nil
}

func (r *PostgresChangeRepository) GetPendingTargetsForChange(
	ctx context.Context,
	changeID uuid.UUID,
) ([]model.ChangeTarget, error) {
	var err error
	rows, err := r.db.Query(ctx, getPendingTargetsForChangeQuery, changeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending targets for change %s: %w", changeID, err)
	}
	defer rows.Close()

	var targets []model.ChangeTarget
	for rows.Next() {
		var target model.ChangeTarget
		err = rows.Scan(&target.Status,
			&target.SyncedAt,
			&target.Server.ID,
			&target.Server.Type,
			&target.Server.HostName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target row: %w", err)
		}
		targets = append(targets, target)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating target rows: %w", err)
	}

	return targets, nil
}

func (r *PostgresChangeRepository) UpdateChangeTargetStatus(
	ctx context.Context,
	changeID uuid.UUID,
	hostname string,
	status model.ChangeTargetStatus,
) error {
	_, err := r.db.Exec(ctx, updateChangeTargetStatusQuery, changeID, hostname, status)
	if err != nil {
		return fmt.Errorf(
			"failed to update change target status for change %s and hostname %s: %w",
			changeID,
			hostname,
			err,
		)
	}
	return nil
}
