package service

import (
	"context"
	"log/slog"

	"cogni-cash/internal/domain/port"
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

func (s *SettingsService) GetAll(ctx context.Context) (map[string]string, error) {
	return s.repo.GetAll(ctx)
}

func (s *SettingsService) UpdateMultiple(ctx context.Context, settings map[string]string) error {
	for key, value := range settings {
		if err := s.repo.Set(ctx, key, value); err != nil {
			s.logger.Error("Failed to update setting", "key", key, "error", err)
			return err
		}
	}
	return nil
}
