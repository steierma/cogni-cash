package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
)

// AuthRepository implements port.AuthRepository using pgx.
type AuthRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

// NewAuthRepository creates a new AuthRepository.
func NewAuthRepository(pool *pgxpool.Pool, logger *slog.Logger) *AuthRepository {
	return &AuthRepository{pool: pool, Logger: logger}
}

func (r *AuthRepository) SaveRefreshToken(ctx context.Context, token entity.RefreshToken) error {
	r.Logger.Info("Saving refresh token", "user_id", token.UserID)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt, token.Revoked)
	if err != nil {
		return fmt.Errorf("auth repo: save refresh token: %w", err)
	}
	return nil
}

func (r *AuthRepository) FindRefreshToken(ctx context.Context, tokenHash string) (entity.RefreshToken, error) {
	r.Logger.Debug("Finding refresh token", "hash", tokenHash)
	var t entity.RefreshToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token_hash = $1`, tokenHash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.Revoked)
	if err != nil {
		return entity.RefreshToken{}, fmt.Errorf("auth repo: find refresh token: %w", err)
	}
	return t, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	r.Logger.Info("Revoking refresh token", "id", id)
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("auth repo: revoke refresh token: %w", err)
	}
	return nil
}

func (r *AuthRepository) RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	r.Logger.Info("Revoking all refresh tokens", "user_id", userID)
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("auth repo: revoke all refresh tokens: %w", err)
	}
	return nil
}

func (r *AuthRepository) CleanupExpiredRefreshTokens(ctx context.Context) error {
	r.Logger.Info("Cleaning up expired refresh tokens")
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE`)
	if err != nil {
		return fmt.Errorf("auth repo: cleanup: %w", err)
	}
	return nil
}
