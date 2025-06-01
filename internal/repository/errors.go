package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEntityNotFound      = errors.New("entity not found")
	ErrEntityAlreadyExists = errors.New("entity already exists")
)

func handleError(err error, msg string, args ...any) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEntityNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrEntityAlreadyExists
		}
	}

	return fmt.Errorf(msg, args...)
}
