// Package postgres provides shared helpers for opening and managing a pgx
// connection pool used by all repository adapters.
package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool opens a *pgxpool.Pool from the given DSN and verifies connectivity
// with a Ping. The caller is responsible for calling pool.Close().
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	slog.Info("Initializing new PostgreSQL connection pool")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("Failed to create connection pool", "error", err)
		return nil, fmt.Errorf("postgres: open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("Failed to ping database on pool initialization", "error", err)
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	slog.Info("PostgreSQL connection pool established and verified successfully")
	return pool, nil
}
