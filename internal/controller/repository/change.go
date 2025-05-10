package repository

import (
	"context"
	"encoding/json"
	"fmt"

	controllerdomain "github.com/davidseybold/beacondns/internal/controller/domain"
	beacondomain "github.com/davidseybold/beacondns/internal/domain"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
	"github.com/google/uuid"
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
		INSERT INTO change_targets (id, change_id, server_id, status)
		VALUES ($1, $2, $3, $4);
	`
)

type ChangeRepository interface {
	CreateChange(ctx context.Context, change beacondomain.Change) (*beacondomain.Change, error)
	GetChange(ctx context.Context, id uuid.UUID) (*beacondomain.Change, error)

	CreateChangeTargets(ctx context.Context, changeTargets []controllerdomain.ChangeTarget) error
}

type PostgresChangeRepository struct {
	db postgres.Queryer
}

func (r *PostgresChangeRepository) CreateChange(ctx context.Context, change beacondomain.Change) (*beacondomain.Change, error) {
	var changeData any
	switch change.Type {
	case beacondomain.ChangeTypeZone:
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

	return &change, nil
}

func (r *PostgresChangeRepository) GetChange(ctx context.Context, id uuid.UUID) (*beacondomain.Change, error) {
	row := r.db.QueryRow(ctx, getChangeQuery, id)

	var change beacondomain.Change
	var changeData []byte
	err := row.Scan(&change.ID, &change.Type, &changeData, &change.SubmittedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get change %s from database: %w", id, err)
	}

	switch change.Type {
	case beacondomain.ChangeTypeZone:
		var zoneChange beacondomain.ZoneChange
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

func (r *PostgresChangeRepository) CreateChangeTargets(ctx context.Context, changeTargets []controllerdomain.ChangeTarget) error {
	for _, changeTarget := range changeTargets {
		_, err := r.db.Exec(ctx, insertChangeTargetQuery, changeTarget.ID, changeTarget.ChangeID, changeTarget.ServerID, changeTarget.Status)
		if err != nil {
			return fmt.Errorf("failed to create change target %s for change %s: %w", changeTarget.ID, changeTarget.ChangeID, err)
		}
	}
	return nil
}
