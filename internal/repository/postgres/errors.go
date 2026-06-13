package postgres

import (
	"Fitness-tracker/internal/repository"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

func mapNotFound(err error, message string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", message, repository.ErrNotFound)
	}
	return err
}

func mapConstraintError(err error, message string) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case uniqueViolation:
		return fmt.Errorf("%s: %w", message, repository.ErrAlreadyExists)
	case foreignKeyViolation:
		return fmt.Errorf("%s: %w", message, repository.ErrNotFound)
	default:
		return err
	}
}

func mapResourceInUse(err error, message string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation {
		return fmt.Errorf("%s: %w", message, repository.ErrResourceInUse)
	}
	return err
}
