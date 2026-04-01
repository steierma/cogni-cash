package service

import (
	"context"
	"log/slog"

	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type SettingsService struct {
	repo   port.SettingsRepository
	logger *slog.Logger
}

func NewSettingsService(repo port.SettingsRepository, logger *slog.Logger) *SettingsService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SettingsService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SettingsService) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	return s.repo.GetAll(ctx, userID)
}

func (s *SettingsService) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	return s.repo.Get(ctx, key, userID)
}

func (s *SettingsService) UpdateMultiple(ctx context.Context, settings map[string]string, userID uuid.UUID) error {
	for key, value := range settings {
		s.logger.Info("Updating setting", "key", key, "value_len", len(value), "user_id", userID)
		if err := s.repo.Set(ctx, key, value, userID); err != nil {
			s.logger.Error("Failed to update setting", "key", key, "user_id", userID, "error", err)
			return err
		}
	}
	return nil
}
