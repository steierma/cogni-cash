package port

import (
	"context"

	"github.com/google/uuid"

	"cogni-cash/internal/domain/entity"
)

// AuthRepository defines the persistence operations for refresh tokens and other auth data.
type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, token entity.RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredRefreshTokens(ctx context.Context) error
}
