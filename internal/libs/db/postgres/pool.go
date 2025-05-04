package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnectionPool(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	connString := createConnectionString(config)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	poolConfig.BeforeConnect = func(c context.Context, cfg *pgx.ConnConfig) error {
		cfg.Password = config.Password

		return nil
	}

	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, PingDB(ctx, pool)
}

func createConnectionString(c Config) string {
	return fmt.Sprintf("host=%s user=%s database=%s port=%d", c.Host, c.User, c.DBName, c.Port)
}

func PingDB(ctx context.Context, db *pgxpool.Pool) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return conn.Ping(ctx)
}
