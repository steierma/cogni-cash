package port

import "context"

// SettingsRepository defines how the application interacts with persistent configuration.
type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	GetAll(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, key string, value string) error
}
