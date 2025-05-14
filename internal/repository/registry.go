package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"github.com/davidseybold/beacondns/internal/db/postgres"
)

// Inspired by this article: https://zenidas.wordpress.com/recipes/repository-pattern-and-transaction-management-in-golang/

type Registry interface {
	GetZoneRepository() ZoneRepository
	GetChangeRepository() ChangeRepository
	GetServerRepository() ServerRepository
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

func (r *PostgresRepositoryRegistry) GetChangeRepository() ChangeRepository {
	db := r.getQueryer()
	return &PostgresChangeRepository{db}
}

func (r *PostgresRepositoryRegistry) GetServerRepository() ServerRepository {
	db := r.getQueryer()
	return &PostgresServerRepository{db}
}

func (r *PostgresRepositoryRegistry) getQueryer() postgres.Queryer {
	if r.queryer != nil {
		return r.queryer
	}
	return r.db
}

func (r *PostgresRepositoryRegistry) WithinTransaction(ctx context.Context, txFunc TxFunc) (any, error) {
	registry := r

	var tx postgres.Tx
	var err error
	if r.queryer == nil {
		tx, err = r.db.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}

		defer func() {
			if p := recover(); p != nil {
				_ = tx.Rollback(ctx)
				panic(p)
			}
		}()

		registry = &PostgresRepositoryRegistry{
			db:      r.db,
			queryer: tx,
		}
	}

	result, err := txFunc(ctx, registry)
	if err != nil {
		if xerr := tx.Rollback(ctx); xerr != nil && !errors.Is(xerr, pgx.ErrTxClosed) {
			return nil, fmt.Errorf("rollback failed: %w; original error: %w", xerr, err)
		}
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}
