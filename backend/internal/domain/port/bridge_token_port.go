package port

import (
	"context"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

// BridgeAccessTokenRepository defines the storage operations for bridge tokens.
type BridgeAccessTokenRepository interface {
	Save(ctx context.Context, token entity.BridgeAccessToken) error
	FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BridgeAccessToken, error)
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BridgeAccessToken, error)
	FindByHash(ctx context.Context, hash string) (entity.BridgeAccessToken, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// BridgeAccessTokenUseCase defines the business logic for bridge tokens.
type BridgeAccessTokenUseCase interface {
	CreateToken(ctx context.Context, userID uuid.UUID, name string) (entity.CreateBridgeTokenResponse, error)
	ListTokens(ctx context.Context, userID uuid.UUID) ([]entity.BridgeAccessToken, error)
	RevokeToken(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)
}
