package database

import (
	"Fitness-tracker/internal/config"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const connectTimeout = 5 * time.Second

func ConnectPostgres(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	connectionString := cfg.ConnectionString()
	if connectionString == "" {
		return nil, errors.New("database connection string is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
