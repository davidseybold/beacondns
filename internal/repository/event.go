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
	insertEventQuery = `
		INSERT INTO events (id, type, payload)
		VALUES ($1, $2, $3)
	`
	getEventWithLockQuery = `
	UPDATE events
		SET lock_expires = NOW() + INTERVAL '5 minutes'
		WHERE id = (
			SELECT id
			FROM events
			WHERE (lock_expires < NOW() OR lock_expires IS NULL)
			ORDER BY created_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, type, payload, created_at
	`

	deleteEventQuery = `
		DELETE FROM events
		WHERE id = $1
	`
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *model.Event) error
	GetEventWithLock(ctx context.Context) (*model.Event, error)
	DeleteEvent(ctx context.Context, id uuid.UUID) error
}

type PostgresEventRepository struct {
	db postgres.Queryer
}

func (r *PostgresEventRepository) CreateEvent(ctx context.Context, event *model.Event) error {
	_, err := r.db.Exec(ctx, insertEventQuery, event.ID, event.Type, event.Payload)
	return err
}

func (r *PostgresEventRepository) GetEventWithLock(ctx context.Context) (*model.Event, error) {
	row := r.db.QueryRow(ctx, getEventWithLockQuery)

	var event model.Event
	err := row.Scan(&event.ID, &event.Type, &event.Payload, &event.CreatedAt)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEntityNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to scan event row: %w", err)
	}

	return &event, nil
}

func (r *PostgresEventRepository) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, deleteEventQuery, id)
	return err
}
