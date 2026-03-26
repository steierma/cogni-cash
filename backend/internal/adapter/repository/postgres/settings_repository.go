package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewSettingsRepository(pool *pgxpool.Pool, logger *slog.Logger) *SettingsRepository {
	return &SettingsRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *SettingsRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx, "SELECT value FROM settings WHERE key = $1", key).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil // Return empty string for unset keys instead of an error
		}
		return "", err
	}
	return value, nil
}

func (r *SettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, rows.Err()
}

func (r *SettingsRepository) Set(ctx context.Context, key string, value string) error {
	query := `
		INSERT INTO settings (key, value) 
		VALUES ($1, $2) 
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`
	_, err := r.pool.Exec(ctx, query, key, value)
	return err
}
