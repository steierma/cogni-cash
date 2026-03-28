package memory

import (
	"context"
	"sync"

	"cogni-cash/internal/domain/port"
)

type SettingsRepository struct {
	mu       sync.RWMutex
	settings map[string]string
}

func NewSettingsRepository() *SettingsRepository {
	return &SettingsRepository{
		settings: make(map[string]string),
	}
}

func (r *SettingsRepository) Get(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, ok := r.settings[key]
	if !ok {
		return "", nil // Return empty string if not found, as per common patterns in this project
	}
	return val, nil
}

func (r *SettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make(map[string]string)
	for k, v := range r.settings {
		res[k] = v
	}
	return res, nil
}

func (r *SettingsRepository) Set(ctx context.Context, key string, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings[key] = value
	return nil
}

var _ port.SettingsRepository = (*SettingsRepository)(nil)
