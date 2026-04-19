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
	vaultKey string
}

func NewSettingsRepository(pool *pgxpool.Pool, userRepo port.UserRepository, vaultKey string, logger *slog.Logger) *SettingsRepository {
	return &SettingsRepository{
		pool:     pool,
		userRepo: userRepo,
		logger:   logger,
		vaultKey: vaultKey,
	}
}

func (r *SettingsRepository) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	var rawBytes []byte
	var isSensitive bool

	query := "SELECT value, is_sensitive FROM settings WHERE key = $1 AND user_id = $2"
	err := r.pool.QueryRow(ctx, query, key, userID).Scan(&rawBytes, &isSensitive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Fallback to admin settings
			adminID, err := r.userRepo.GetAdminID(ctx)
			if err != nil || adminID == userID {
				return "", nil
			}
			err = r.pool.QueryRow(ctx, query, key, adminID).Scan(&rawBytes, &isSensitive)
			if err != nil {
				return "", nil // Not found for admin either
			}
		} else {
			return "", err
		}
	}

	if !isSensitive {
		return string(rawBytes), nil
	}

	// Try decryption
	var decrypted string
	err = r.pool.QueryRow(ctx, "SELECT convert_from(pgp_sym_decrypt_bytea($1, $2), 'utf8')", rawBytes, r.vaultKey).Scan(&decrypted)
	if err != nil {
		return string(rawBytes), nil // Fallback
	}
	return decrypted, nil
}

func (r *SettingsRepository) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT key, value, is_sensitive FROM settings WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key string
		var rawBytes []byte
		var isSensitive bool
		if err := rows.Scan(&key, &rawBytes, &isSensitive); err != nil {
			return nil, err
		}

		if !isSensitive {
			settings[key] = string(rawBytes)
			continue
		}

		// Try decryption in DB
		var decrypted string
		err := r.pool.QueryRow(ctx, "SELECT convert_from(pgp_sym_decrypt_bytea($1, $2), 'utf8')", rawBytes, r.vaultKey).Scan(&decrypted)
		if err != nil {
			// Fallback to raw value on error (e.g. legacy plain text)
			settings[key] = string(rawBytes)
		} else {
			settings[key] = decrypted
		}
	}
	return settings, rows.Err()
}

func (r *SettingsRepository) Set(ctx context.Context, key string, value string, userID uuid.UUID, isSensitive bool) error {
	var query string
	var args []interface{}

	if isSensitive {
		query = `
			INSERT INTO settings (key, value, user_id, is_sensitive) 
			VALUES ($1, pgp_sym_encrypt_bytea($2, $3), $4, $5) 
			ON CONFLICT (key, user_id) DO UPDATE SET value = EXCLUDED.value, is_sensitive = EXCLUDED.is_sensitive
		`
		args = []interface{}{key, value, r.vaultKey, userID, true}
	} else {
		query = `
			INSERT INTO settings (key, value, user_id, is_sensitive) 
			VALUES ($1, $2, $3, $4) 
			ON CONFLICT (key, user_id) DO UPDATE SET value = EXCLUDED.value, is_sensitive = EXCLUDED.is_sensitive
		`
		args = []interface{}{key, []byte(value), userID, false}
	}

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}
