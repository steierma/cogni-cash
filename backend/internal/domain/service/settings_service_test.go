package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

// --- Mock Settings Repository ---

type mockSimpleSettingsRepo struct {
	data   map[string]string
	setErr error
	getErr error
}

func newMockSimpleSettingsRepo() *mockSimpleSettingsRepo {
	return &mockSimpleSettingsRepo{data: make(map[string]string)}
}

func (m *mockSimpleSettingsRepo) Get(_ context.Context, key string, _ uuid.UUID) (string, error) {
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return "", errors.New("key not found")
}

func (m *mockSimpleSettingsRepo) GetGlobal(_ context.Context, key string) (string, error) {
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return "", errors.New("key not found")
}

func (m *mockSimpleSettingsRepo) GetAll(_ context.Context, _ uuid.UUID) (map[string]string, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	cp := make(map[string]string, len(m.data))
	for k, v := range m.data {
		cp[k] = v
	}
	return cp, nil
}

func (m *mockSimpleSettingsRepo) Set(_ context.Context, key string, value string, _ uuid.UUID, isSensitive bool) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}

// --- Tests ---

func TestSettingsService_GetAllMasked_Admin(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	repo.data["theme"] = "dark"
	repo.data["smtp_password"] = "super-secret"
	repo.data["llm_api_token"] = "token-123"

	svc := service.NewSettingsService(repo, setupLogger())
	// Teste als Admin (true)
	settings, err := svc.GetAllMasked(context.Background(), uuid.New(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if settings["theme"] != "dark" {
		t.Errorf("expected theme 'dark', got %q", settings["theme"])
	}
	if settings["smtp_password"] != "********" {
		t.Errorf("expected smtp_password masked, got %q", settings["smtp_password"])
	}
	if settings["llm_api_token"] != "********" {
		t.Errorf("expected llm_api_token masked, got %q", settings["llm_api_token"])
	}
}

func TestSettingsService_GetAllMasked_NonAdmin_HidesKeys(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	repo.data["theme"] = "light"
	repo.data["smtp_host"] = "smtp.example.com"
	repo.data["llm_model"] = "gpt-4"

	svc := service.NewSettingsService(repo, setupLogger())
	// Teste als normaler User (false)
	settings, err := svc.GetAllMasked(context.Background(), uuid.New(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if settings["theme"] != "light" {
		t.Errorf("expected user setting 'theme' to be visible, got %q", settings["theme"])
	}
	if _, exists := settings["smtp_host"]; exists {
		t.Errorf("expected admin setting 'smtp_host' to be hidden from non-admin")
	}
	if _, exists := settings["llm_model"]; exists {
		t.Errorf("expected admin setting 'llm_model' to be hidden from non-admin")
	}
}

func TestSettingsService_UpdateMultiple_IgnoreMask(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	repo.data["smtp_password"] = "original-secret"
	svc := service.NewSettingsService(repo, setupLogger())

	err := svc.UpdateMultiple(context.Background(), map[string]string{
		"theme":         "light",
		"smtp_password": "********",
	}, uuid.New(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.data["theme"] != "light" {
		t.Errorf("expected theme 'light', got %q", repo.data["theme"])
	}
	if repo.data["smtp_password"] != "original-secret" {
		t.Errorf("expected smtp_password to remain 'original-secret', got %q", repo.data["smtp_password"])
	}
}

func TestSettingsService_Get(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	repo.data["theme"] = "dark"
	svc := service.NewSettingsService(repo, setupLogger())

	val, err := svc.Get(context.Background(), "theme", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "dark" {
		t.Errorf("expected 'dark', got %q", val)
	}
}

func TestSettingsService_GetAllMasked_RepoError(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	repo.getErr = errors.New("db error")
	svc := service.NewSettingsService(repo, setupLogger())

	_, err := svc.GetAllMasked(context.Background(), uuid.New(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSettingsService_UpdateMultiple_RepoError(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	repo.setErr = errors.New("db error")
	svc := service.NewSettingsService(repo, setupLogger())

	err := svc.UpdateMultiple(context.Background(), map[string]string{"theme": "light"}, uuid.New(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSettingsService_GetAll(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	repo.data["theme"] = "dark"
	repo.data["currency"] = "EUR"

	svc := service.NewSettingsService(repo, setupLogger())
	settings, err := svc.GetAll(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(settings) != 2 {
		t.Errorf("expected 2 settings, got %d", len(settings))
	}
	if settings["theme"] != "dark" {
		t.Errorf("expected theme 'dark', got %q", settings["theme"])
	}
}

func TestSettingsService_GetAll_Empty(t *testing.T) {
	svc := service.NewSettingsService(newMockSimpleSettingsRepo(), setupLogger())
	settings, err := svc.GetAll(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(settings) != 0 {
		t.Errorf("expected 0 settings, got %d", len(settings))
	}
}

func TestSettingsService_UpdateMultiple_Success(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	svc := service.NewSettingsService(repo, setupLogger())

	err := svc.UpdateMultiple(context.Background(), map[string]string{
		"theme":    "dark",
		"currency": "USD",
	}, uuid.New(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.data["theme"] != "dark" {
		t.Errorf("expected theme 'dark', got %q", repo.data["theme"])
	}
	if repo.data["currency"] != "USD" {
		t.Errorf("expected currency 'USD', got %q", repo.data["currency"])
	}
}

func TestSettingsService_UpdateMultiple_NonAdmin_Restriction(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	svc := service.NewSettingsService(repo, setupLogger())

	err := svc.UpdateMultiple(context.Background(), map[string]string{
		"theme":     "dark",
		"smtp_host": "malicious-smtp.com",
	}, uuid.New(), false) // isAdmin = false
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.data["theme"] != "dark" {
		t.Errorf("expected theme 'dark' to be updated, got %q", repo.data["theme"])
	}
	if _, exists := repo.data["smtp_host"]; exists {
		t.Errorf("expected smtp_host to be ignored for non-admin")
	}
}

func TestSettingsService_UpdateMultiple_Error(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	repo.setErr = errors.New("db write failed")

	svc := service.NewSettingsService(repo, setupLogger())

	err := svc.UpdateMultiple(context.Background(), map[string]string{
		"theme": "dark",
	}, uuid.New(), true)
	if err == nil {
		t.Error("expected error when Set fails")
	}
}

func TestSettingsService_LLMProfiles_MaskingAndMerging(t *testing.T) {
	repo := newMockSimpleSettingsRepo()
	originalProfiles := `[{"id":"1","name":"Profile 1","token":"secret-1"},{"id":"2","name":"Profile 2","token":"secret-2"}]`
	repo.data["llm_profiles"] = originalProfiles

	svc := service.NewSettingsService(repo, setupLogger())
	userID := uuid.New()

	t.Run("Masks tokens in GetAllMasked", func(t *testing.T) {
		settings, err := svc.GetAllMasked(context.Background(), userID, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		val := settings["llm_profiles"]
		if !strings.Contains(val, "********") {
			t.Errorf("expected masked tokens, got %s", val)
		}
		if strings.Contains(val, "secret-1") {
			t.Errorf("expected secret-1 to be hidden, got %s", val)
		}
	})

	t.Run("Merges tokens in UpdateMultiple", func(t *testing.T) {
		newProfiles := `[{"id":"1","name":"Profile 1 Updated","token":"********"},{"id":"2","name":"Profile 2","token":"new-secret-2"},{"id":"3","name":"New Profile","token":"secret-3"}]`
		
		err := svc.UpdateMultiple(context.Background(), map[string]string{
			"llm_profiles": newProfiles,
		}, userID, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		savedJSON := repo.data["llm_profiles"]
		if !strings.Contains(savedJSON, "secret-1") {
			t.Errorf("expected secret-1 to be preserved for profile 1, got %s", savedJSON)
		}
		if !strings.Contains(savedJSON, "new-secret-2") {
			t.Errorf("expected new-secret-2 to be saved for profile 2, got %s", savedJSON)
		}
		if !strings.Contains(savedJSON, "secret-3") {
			t.Errorf("expected secret-3 to be saved for new profile 3, got %s", savedJSON)
		}
		if strings.Contains(savedJSON, "********") {
			t.Errorf("saved JSON should not contain asterisks, got %s", savedJSON)
		}
	})
}

func TestSettingsService_NilLogger(t *testing.T) {
	// Should not panic with nil logger
	svc := service.NewSettingsService(newMockSimpleSettingsRepo(), nil)
	_, err := svc.GetAll(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
