package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// PasswordResetRepository defines how password reset tokens are persisted.
type PasswordResetRepository interface {
	Create(ctx context.Context, token entity.PasswordResetToken) error
	FindByHash(ctx context.Context, tokenHash string) (entity.PasswordResetToken, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	CleanupExpired(ctx context.Context) error
}
