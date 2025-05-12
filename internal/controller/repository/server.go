package repository

import (
	"context"
	"fmt"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
	"github.com/jackc/pgx/v5"
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
	CreateServer(ctx context.Context, server *domain.Server) error
	GetServerByHostName(ctx context.Context, hostName string) (*domain.Server, error)
	GetAllServers(ctx context.Context) ([]*domain.Server, error)
}

type PostgresServerRepository struct {
	db postgres.Queryer
}

func (r *PostgresServerRepository) CreateServer(ctx context.Context, server *domain.Server) error {
	_, err := r.db.Exec(ctx, insertServerQuery, server.ID, server.HostName)
	if err != nil {
		return fmt.Errorf("failed to create server %s: %w", server.HostName, err)
	}
	return nil
}

func (r *PostgresServerRepository) GetServerByHostName(ctx context.Context, hostName string) (*domain.Server, error) {
	rows := r.db.QueryRow(ctx, getServerByHostNameQuery, hostName)
	server, err := scanServer(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to get server by host name %s: %w", hostName, err)
	}
	return server, nil
}

func (r *PostgresServerRepository) GetAllServers(ctx context.Context) ([]*domain.Server, error) {
	rows, err := r.db.Query(ctx, getAllServersQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query all servers: %w", err)
	}
	defer rows.Close()

	servers := make([]*domain.Server, 0)
	for rows.Next() {
		server, err := scanServer(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server row: %w", err)
		}
		servers = append(servers, server)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating server rows: %w", err)
	}
	return servers, nil
}

func scanServer(rows pgx.Row) (*domain.Server, error) {
	var server domain.Server
	err := rows.Scan(&server.ID, &server.Type, &server.HostName)
	if err != nil {
		return nil, fmt.Errorf("failed to scan server row: %w", err)
	}
	return &server, nil
}
