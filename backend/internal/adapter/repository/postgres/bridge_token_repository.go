package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BridgeAccessTokenRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

func NewBridgeAccessTokenRepository(pool *pgxpool.Pool, logger *slog.Logger) *BridgeAccessTokenRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &BridgeAccessTokenRepository{pool: pool, Logger: logger}
}

func (r *BridgeAccessTokenRepository) Save(ctx context.Context, token entity.BridgeAccessToken) error {
	r.Logger.Info("Saving bridge access token", "user_id", token.UserID, "name", token.Name)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bridge_access_tokens (id, user_id, name, token_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		token.ID, token.UserID, token.Name, token.TokenHash, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("bridge_token repo: save: %w", err)
	}
	return nil
}

func (r *BridgeAccessTokenRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BridgeAccessToken, error) {
	r.Logger.Debug("Finding all bridge tokens", "user_id", userID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, token_hash, last_used_at, created_at
		FROM bridge_access_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("bridge_token repo: find all: %w", err)
	}
	defer rows.Close()

	var tokens []entity.BridgeAccessToken
	for rows.Next() {
		var t entity.BridgeAccessToken
		err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.TokenHash, &t.LastUsedAt, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("bridge_token repo: scan: %w", err)
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func (r *BridgeAccessTokenRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BridgeAccessToken, error) {
	r.Logger.Debug("Finding bridge token by ID", "id", id, "user_id", userID)
	var t entity.BridgeAccessToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, token_hash, last_used_at, created_at
		FROM bridge_access_tokens
		WHERE id = $1 AND user_id = $2`, id, userID).
		Scan(&t.ID, &t.UserID, &t.Name, &t.TokenHash, &t.LastUsedAt, &t.CreatedAt)
	if err != nil {
		return entity.BridgeAccessToken{}, fmt.Errorf("bridge_token repo: find by id: %w", err)
	}
	return t, nil
}

func (r *BridgeAccessTokenRepository) FindByHash(ctx context.Context, hash string) (entity.BridgeAccessToken, error) {
	r.Logger.Debug("Finding bridge token by hash")
	var t entity.BridgeAccessToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, token_hash, last_used_at, created_at
		FROM bridge_access_tokens
		WHERE token_hash = $1`, hash).
		Scan(&t.ID, &t.UserID, &t.Name, &t.TokenHash, &t.LastUsedAt, &t.CreatedAt)
	if err != nil {
		return entity.BridgeAccessToken{}, fmt.Errorf("bridge_token repo: find by hash: %w", err)
	}
	return t, nil
}

func (r *BridgeAccessTokenRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.Logger.Debug("Updating bridge token last_used_at", "id", id, "user_id", userID)
	_, err := r.pool.Exec(ctx, `
		UPDATE bridge_access_tokens
		SET last_used_at = NOW()
		WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("bridge_token repo: update last_used: %w", err)
	}
	return nil
}

func (r *BridgeAccessTokenRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.Logger.Info("Deleting bridge token", "id", id, "user_id", userID)
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM bridge_access_tokens
		WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("bridge_token repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bridge_token repo: delete: not found")
	}
	return nil
}
