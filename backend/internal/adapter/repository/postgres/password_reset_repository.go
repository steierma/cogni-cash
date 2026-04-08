package postgres

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPasswordResetRepository(pool *pgxpool.Pool, logger *slog.Logger) *PasswordResetRepository {
	return &PasswordResetRepository{
		pool:   pool,
		logger: logger,
	}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token entity.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	return err
}

func (r *PasswordResetRepository) FindByHash(ctx context.Context, tokenHash string) (entity.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	var t entity.PasswordResetToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.PasswordResetToken{}, entity.ErrResetTokenInvalid
		}
		return entity.PasswordResetToken{}, err
	}
	return t, nil
}

func (r *PasswordResetRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := "DELETE FROM password_reset_tokens WHERE user_id = $1"
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *PasswordResetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM password_reset_tokens WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PasswordResetRepository) CleanupExpired(ctx context.Context) error {
	query := "DELETE FROM password_reset_tokens WHERE expires_at < NOW()"
	_, err := r.pool.Exec(ctx, query)
	return err
}
