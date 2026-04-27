package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const maskValue = "********"

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

func (s *SettingsService) GetAllMasked(ctx context.Context, userID uuid.UUID, isAdmin bool) (map[string]string, error) {
	settings, err := s.repo.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	masked := make(map[string]string)
	for k, v := range settings {
		// Filter out admin-only keys completely for non-admin users
		if adminOnlyKeys[k] && !isAdmin {
			continue
		}

		if k == "llm_profiles" && v != "" {
			var profiles []entity.LLMProfile
			if err := json.Unmarshal([]byte(v), &profiles); err == nil {
				for i := range profiles {
					if profiles[i].Token != "" {
						profiles[i].Token = maskValue
					}
				}
				if maskedJSON, err := json.Marshal(profiles); err == nil {
					masked[k] = string(maskedJSON)
					continue
				}
			}
		}

		if s.isSensitiveKey(k) && v != "" {
			masked[k] = maskValue
		} else {
			masked[k] = v
		}
	}
	return masked, nil
}

func (s *SettingsService) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	return s.repo.Get(ctx, key, userID)
}

var adminOnlyKeys = map[string]bool{
	"smtp_host":                      true,
	"smtp_port":                      true,
	"smtp_user":                      true,
	"smtp_password":                  true,
	"smtp_from_email":                true,
	"llm_api_url":                    true,
	"llm_api_token":                  true,
	"llm_model":                      true,
	"llm_single_prompt":              true,
	"llm_batch_prompt":               true,
	"llm_statement_prompt":           true,
	"llm_payslip_prompt":             true,
	"import_dir":                     true,
	"import_interval":                true,
	"auto_categorization_enabled":    true,
	"auto_categorization_interval":   true,
	"auto_categorization_batch_size": true,
	"auto_categorization_examples_per_category": true,
	"bank_provider":            true,
	"enablebanking_app_id":     true,
	"bank_sync_enabled":        true,
	"bank_sync_interval":       true,
	"bank_sync_next_run":       true,
	"payslip_import_json_path": true,
	"payslip_import_interval":  true,
	"llm_enforce_user_config":  true,
	"llm_subscription_prompt":  true,
	"llm_cancellation_prompt":  true,
}

func (s *SettingsService) UpdateMultiple(ctx context.Context, settings map[string]string, userID uuid.UUID, isAdmin bool) error {
	for key, value := range settings {
		// If value is the mask, it means the frontend hasn't changed it (it just sent back what it got)
		// We skip updating in this case to avoid overwriting the real secret with asterisks.
		if value == maskValue && s.isSensitiveKey(key) {
			continue
		}

		// Role-based validation
		if adminOnlyKeys[key] && !isAdmin {
			s.logger.Warn("Non-admin user attempted to update restricted setting", "user_id", userID, "key", key)
			continue // Or return error? The frontend already hides these, so a "continue" is a silent guard.
		}

		// Special handling for llm_profiles JSON merging
		if key == "llm_profiles" && value != "" {
			var newProfiles []entity.LLMProfile
			if err := json.Unmarshal([]byte(value), &newProfiles); err == nil {
				// Fetch existing to merge tokens
				existingJSON, _ := s.repo.Get(ctx, key, userID)
				if existingJSON != "" {
					var oldProfiles []entity.LLMProfile
					if err := json.Unmarshal([]byte(existingJSON), &oldProfiles); err == nil {
						// Create map for easy lookup
						oldMap := make(map[string]entity.LLMProfile)
						for _, p := range oldProfiles {
							oldMap[p.ID] = p
						}

						// Merge
						for i := range newProfiles {
							if newProfiles[i].Token == maskValue {
								if old, ok := oldMap[newProfiles[i].ID]; ok {
									newProfiles[i].Token = old.Token
								} else {
									newProfiles[i].Token = "" // Should not happen if masked
								}
							}
						}
					}
				}
				// Marshal back to save
				if mergedJSON, err := json.Marshal(newProfiles); err == nil {
					value = string(mergedJSON)
				}
			}
		}

		isSensitive := s.isSensitiveKey(key)
		s.logger.Info("Updating setting", "key", key, "value_len", len(value), "user_id", userID, "is_sensitive", isSensitive)
		if err := s.repo.Set(ctx, key, value, userID, isSensitive); err != nil {
			s.logger.Error("Failed to update setting", "key", key, "user_id", userID, "error", err)
			return err
		}
	}
	return nil
}

func (s *SettingsService) isSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "password") ||
		strings.Contains(k, "token") ||
		strings.Contains(k, "secret") ||
		strings.Contains(k, "key")
}
