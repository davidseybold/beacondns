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
	GetResponsePolicyRepository() ResponsePolicyRepository
	GetEventRepository() EventRepository
}

type Transactor interface {
	InTx(ctx context.Context, txFunc TxFunc) error
}

type TransactorRegistry interface {
	Transactor
	Registry
}

type TxFunc func(ctx context.Context, r Registry) error

type PostgresRepositoryRegistry struct {
	db      postgres.PgxPool
	queryer postgres.Queryer
}

var _ TransactorRegistry = (*PostgresRepositoryRegistry)(nil)

func NewPostgresRepositoryRegistry(db postgres.PgxPool) *PostgresRepositoryRegistry {
	return &PostgresRepositoryRegistry{
		db: db,
	}
}

func (r *PostgresRepositoryRegistry) GetZoneRepository() ZoneRepository {
	db := r.getQueryer()
	return &PostgresZoneRepository{db}
}

func (r *PostgresRepositoryRegistry) GetResponsePolicyRepository() ResponsePolicyRepository {
	db := r.getQueryer()
	return &PostgresResponsePolicyRepository{db}
}

func (r *PostgresRepositoryRegistry) GetEventRepository() EventRepository {
	db := r.getQueryer()
	return &PostgresEventRepository{db}
}

func (r *PostgresRepositoryRegistry) getQueryer() postgres.Queryer {
	if r.queryer != nil {
		return r.queryer
	}
	return r.db
}

func (r *PostgresRepositoryRegistry) InTx(ctx context.Context, txFunc TxFunc) error {
	registry := r

	var tx postgres.Tx
	var err error
	if r.queryer == nil {
		tx, err = r.db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
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

	err = txFunc(ctx, registry)
	if err != nil {
		if xerr := tx.Rollback(ctx); xerr != nil && !errors.Is(xerr, pgx.ErrTxClosed) {
			return fmt.Errorf("rollback failed: %w; original error: %w", xerr, err)
		}
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
