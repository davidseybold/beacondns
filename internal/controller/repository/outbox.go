package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
)

const (
	insertOutboxMessageQuery = `
	INSERT INTO outbox (id, route_key, payload)
	VALUES ($1, $2, $3);
	`
)

type OutboxRepository interface {
	GetPendingMessages(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error
	InsertMessages(ctx context.Context, msgs []domain.OutboxMessage) error
}

type PostgresOutboxRepository struct {
	db postgres.Queryer
}

func (p *PostgresOutboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]domain.OutboxMessage, error) {
	rows, err := p.db.Query(ctx, "SELECT id, payload, route_key FROM outbox WHERE status = 'pending' ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.OutboxMessage
	for rows.Next() {
		var msg domain.OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.Payload, &msg.RouteKey); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (p *PostgresOutboxRepository) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	_, err := p.db.Exec(ctx, "DELETE FROM outbox WHERE id = $1", id)
	return err
}

func (p *PostgresOutboxRepository) InsertMessages(ctx context.Context, msgs []domain.OutboxMessage) error {
	batch := &pgx.Batch{}

	for _, m := range msgs {
		batch.Queue(insertOutboxMessageQuery, m.ID, m.RouteKey, m.Payload)
	}

	results := p.db.SendBatch(ctx, batch)
	defer results.Close()

	for _, _ = range msgs {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("failed to insert outbox message: %w", err)
		}
	}

	return nil
}
