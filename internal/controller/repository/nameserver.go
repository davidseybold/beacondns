package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
)

const (
	insertNameserverQuery = "INSERT INTO nameservers (id, name, routing_key) VALUES ($1, $2, $3) RETURNING id, name, routing_key;"
	listNameserversQuery  = "SELECT id, name, routing_key FROM nameservers ORDER BY name;"

	selectRandomNameServersQuery = "SELECT id, name, routing_key FROM nameservers ORDER BY RANDOM() LIMIT $1;"
)

type NameServerRepository interface {
	AddNameServer(ctx context.Context, ns *domain.NameServer) (*domain.NameServer, error)
	ListNameServers(ctx context.Context) ([]domain.NameServer, error)
	GetRandomNameServers(ctx context.Context, count int) ([]domain.NameServer, error)
}

type PostgresNameServerRepository struct {
	db postgres.Queryer
}

func NewPostgresNameServerRepository() *PostgresNameServerRepository {
	return &PostgresNameServerRepository{}
}

func (p *PostgresNameServerRepository) AddNameServer(ctx context.Context, ns *domain.NameServer) (*domain.NameServer, error) {
	row := p.db.QueryRow(ctx, insertNameserverQuery, ns.ID, ns.Name, ns.RouteKey)

	dbNs, err := scanNameServer(row)
	if err != nil {
		return nil, err
	}

	return dbNs, nil
}

func (p *PostgresNameServerRepository) ListNameServers(ctx context.Context) ([]domain.NameServer, error) {
	rows, err := p.db.Query(ctx, listNameserversQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNameServers(rows)
}

func (p *PostgresNameServerRepository) GetRandomNameServers(ctx context.Context, count int) ([]domain.NameServer, error) {
	rows, err := p.db.Query(ctx, selectRandomNameServersQuery, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNameServers(rows)
}

func scanNameServers(rows pgx.Rows) ([]domain.NameServer, error) {
	nameServers := make([]domain.NameServer, 0)
	for rows.Next() {
		ns, err := scanNameServer(rows)
		if err != nil {
			return nil, err
		}
		nameServers = append(nameServers, *ns)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nameServers, nil
}

func scanNameServer(row pgx.Row) (*domain.NameServer, error) {
	var ns domain.NameServer
	if err := row.Scan(&ns.ID, &ns.Name, &ns.RouteKey); err != nil {
		return nil, err
	}

	return &ns, nil
}
