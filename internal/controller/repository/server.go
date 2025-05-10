package repository

import (
	"context"
	"fmt"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
)

const (
	insertServerQuery = `
		INSERT INTO servers (id, name)
		VALUES ($1, $2)
	`

	getServerByNameQuery = `
		SELECT id, name
		FROM servers
		WHERE name = $1
	`

	getAllServersQuery = `
		SELECT id, name
		FROM servers
	`
)

type ServerRepository interface {
	CreateServer(ctx context.Context, server *domain.Server) error
	GetServerByName(ctx context.Context, name string) (*domain.Server, error)
	GetAllServers(ctx context.Context) ([]*domain.Server, error)
}

type PostgresServerRepository struct {
	db postgres.Queryer
}

func (r *PostgresServerRepository) CreateServer(ctx context.Context, server *domain.Server) error {
	_, err := r.db.Exec(ctx, insertServerQuery, server.ID, server.Name)
	if err != nil {
		return fmt.Errorf("failed to create server %s: %w", server.Name, err)
	}
	return nil
}

func (r *PostgresServerRepository) GetServerByName(ctx context.Context, name string) (*domain.Server, error) {
	var server domain.Server
	err := r.db.QueryRow(ctx, getServerByNameQuery, name).Scan(&server.ID, &server.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get server by name %s: %w", name, err)
	}
	return &server, nil
}

func (r *PostgresServerRepository) GetAllServers(ctx context.Context) ([]*domain.Server, error) {
	rows, err := r.db.Query(ctx, getAllServersQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query all servers: %w", err)
	}
	defer rows.Close()

	servers := make([]*domain.Server, 0)
	for rows.Next() {
		var server domain.Server
		err := rows.Scan(&server.ID, &server.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server row: %w", err)
		}
		servers = append(servers, &server)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating server rows: %w", err)
	}
	return servers, nil
}
