package postgres

import (
	"context"
	"errors"
	"log/slog"

	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsRepository struct {
	pool     *pgxpool.Pool
	userRepo port.UserRepository
	logger   *slog.Logger
}

func NewSettingsRepository(pool *pgxpool.Pool, userRepo port.UserRepository, logger *slog.Logger) *SettingsRepository {
	return &SettingsRepository{
		pool:     pool,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (r *SettingsRepository) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx, "SELECT value FROM settings WHERE key = $1 AND user_id = $2", key, userID).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Fallback to admin settings if not found for user
			adminID, err := r.userRepo.GetAdminID(ctx)
			if err != nil || adminID == userID {
				return "", nil
			}
			err = r.pool.QueryRow(ctx, "SELECT value FROM settings WHERE key = $1 AND user_id = $2", key, adminID).Scan(&value)
			if err != nil {
				return "", nil // No fallback found either
			}
			return value, nil
		}
		return "", err
	}
	return value, nil
}

func (r *SettingsRepository) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT key, value FROM settings WHERE user_id = $1", userID)
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

func (r *SettingsRepository) Set(ctx context.Context, key string, value string, userID uuid.UUID) error {
	query := `
		INSERT INTO settings (key, value, user_id) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (key, user_id) DO UPDATE SET value = EXCLUDED.value
	`
	_, err := r.pool.Exec(ctx, query, key, value, userID)
	return err
}
