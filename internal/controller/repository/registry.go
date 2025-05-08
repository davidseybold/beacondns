package repository

import (
	"context"
	"fmt"

	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

// Inspired by this article: https://zenidas.wordpress.com/recipes/repository-pattern-and-transaction-management-in-golang/

type Registry interface {
	GetZoneRepository() ZoneRepository
	GetOutboxRepository() OutboxRepository
}

type Transactor interface {
	WithinTransaction(ctx context.Context, txFunc TxFunc) (any, error)
}

type TransactorRegistry interface {
	Transactor
	Registry
}

type TxFunc func(ctx context.Context, r Registry) (any, error)

type PostgresRepositoryRegistry struct {
	db      postgres.PgxPool
	queryer postgres.Queryer
}

func NewPostgresRepositoryRegistry(db postgres.PgxPool) *PostgresRepositoryRegistry {
	return &PostgresRepositoryRegistry{
		db: db,
	}
}

func (r *PostgresRepositoryRegistry) GetZoneRepository() ZoneRepository {
	db := r.getQueryer()
	return &PostgresZoneRepository{db}
}

func (r *PostgresRepositoryRegistry) GetOutboxRepository() OutboxRepository {
	db := r.getQueryer()
	return &PostgresOutboxRepository{db}
}

func (r *PostgresRepositoryRegistry) getQueryer() postgres.Queryer {
	if r.queryer != nil {
		return r.queryer
	}
	return r.db
}

func (r *PostgresRepositoryRegistry) WithinTransaction(ctx context.Context, txFunc TxFunc) (out any, err error) {
	registry := r

	var tx postgres.Tx
	if r.queryer == nil {
		tx, err = r.db.Begin(ctx)

		defer func() {
			if p := recover(); p != nil {
				_ = tx.Rollback(ctx)
				panic(p)
			}
			if err != nil {
				if xerr := tx.Rollback(ctx); xerr != nil && !errors.Is(xerr, pgx.ErrTxClosed) {
					err = fmt.Errorf("rollback failed: %v; original error: %w", xerr, err)
				}
				return
			}
			err = tx.Commit(ctx)
		}()

		registry = &PostgresRepositoryRegistry{
			db:      r.db,
			queryer: tx,
		}
	}

	out, err = txFunc(ctx, registry)

	return
}
