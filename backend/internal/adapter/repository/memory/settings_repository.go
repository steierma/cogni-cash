package memory

import (
	"context"
	"sync"

	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type settingValue struct {
	value       string
	isSensitive bool
}

type SettingsRepository struct {
	mu       sync.RWMutex
	settings map[uuid.UUID]map[string]settingValue
	userRepo port.UserRepository
}

func NewSettingsRepository(userRepo port.UserRepository) *SettingsRepository {
	return &SettingsRepository{
		settings: make(map[uuid.UUID]map[string]settingValue),
		userRepo: userRepo,
	}
}

func (r *SettingsRepository) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	r.mu.RLock()
	val, ok := r.getSetting(userID, key)
	r.mu.RUnlock()

	if ok && val != "" {
		return val, nil
	}

	// Fallback to admin settings
	adminID, err := r.userRepo.GetAdminID(ctx)
	if err != nil || adminID == userID {
		return val, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	adminVal, _ := r.getSetting(adminID, key)
	if adminVal != "" {
		return adminVal, nil
	}

	return val, nil
}

func (r *SettingsRepository) getSetting(userID uuid.UUID, key string) (string, bool) {
	userSettings, ok := r.settings[userID]
	if !ok {
		return "", false
	}
	sVal, ok := userSettings[key]
	return sVal.value, ok
}

func (r *SettingsRepository) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make(map[string]string)
	userSettings, ok := r.settings[userID]
	if !ok {
		return res, nil
	}
	for k, v := range userSettings {
		res[k] = v.value
	}
	return res, nil
}

func (r *SettingsRepository) Set(ctx context.Context, key string, value string, userID uuid.UUID, isSensitive bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.settings[userID]; !ok {
		r.settings[userID] = make(map[string]settingValue)
	}
	r.settings[userID][key] = settingValue{
		value:       value,
		isSensitive: isSensitive,
	}
	return nil
}

var _ port.SettingsRepository = (*SettingsRepository)(nil)
