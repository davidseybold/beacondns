package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/davidseybold/beacondns/internal/db/postgres"
	"github.com/davidseybold/beacondns/internal/model"
)

const (
	insertServerQuery = `
		INSERT INTO servers (id, type, hostname)
		VALUES ($1, $2, $3)
	`

	getServerByHostNameQuery = `
		SELECT id, type, hostname
		FROM servers
		WHERE hostname = $1
	`

	getAllServersQuery = `
		SELECT id, type, hostname
		FROM servers
	`
)

type ServerRepository interface {
	CreateServer(ctx context.Context, server *model.Server) error
	GetServerByHostName(ctx context.Context, hostName string) (*model.Server, error)
	GetAllServers(ctx context.Context) ([]*model.Server, error)
}

type PostgresServerRepository struct {
	db postgres.Queryer
}

func (r *PostgresServerRepository) CreateServer(ctx context.Context, server *model.Server) error {
	_, err := r.db.Exec(ctx, insertServerQuery, server.ID, server.HostName)
	if err != nil {
		return fmt.Errorf("failed to create server %s: %w", server.HostName, err)
	}
	return nil
}

func (r *PostgresServerRepository) GetServerByHostName(ctx context.Context, hostName string) (*model.Server, error) {
	rows := r.db.QueryRow(ctx, getServerByHostNameQuery, hostName)
	server, err := scanServer(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to get server by host name %s: %w", hostName, err)
	}
	return server, nil
}

func (r *PostgresServerRepository) GetAllServers(ctx context.Context) ([]*model.Server, error) {
	var err error
	rows, err := r.db.Query(ctx, getAllServersQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query all servers: %w", err)
	}
	defer rows.Close()

	servers := make([]*model.Server, 0)
	for rows.Next() {
		server, scanErr := scanServer(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan server row: %w", scanErr)
		}
		servers = append(servers, server)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating server rows: %w", err)
	}
	return servers, nil
}

func scanServer(rows pgx.Row) (*model.Server, error) {
	var server model.Server
	err := rows.Scan(&server.ID, &server.Type, &server.HostName)
	if err != nil {
		return nil, fmt.Errorf("failed to scan server row: %w", err)
	}
	return &server, nil
}
