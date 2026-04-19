package port

import (
	"context"

	"github.com/google/uuid"
)

// SettingsRepository defines how the application interacts with persistent configuration.
type SettingsRepository interface {
	Get(ctx context.Context, key string, userID uuid.UUID) (string, error)
	GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error)
	Set(ctx context.Context, key string, value string, userID uuid.UUID, isSensitive bool) error
}
