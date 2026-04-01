package service_test

import (
	"context"
	"errors"
	"testing"

	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

// --- Mock Settings Repository ---

type mockSettingsRepo struct {
	data   map[string]string
	setErr error
}

func newMockSettingsRepo() *mockSettingsRepo {
	return &mockSettingsRepo{data: make(map[string]string)}
}

func (m *mockSettingsRepo) Get(_ context.Context, key string, _ uuid.UUID) (string, error) {
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return "", errors.New("key not found")
}

func (m *mockSettingsRepo) GetAll(_ context.Context, _ uuid.UUID) (map[string]string, error) {
	cp := make(map[string]string, len(m.data))
	for k, v := range m.data {
		cp[k] = v
	}
	return cp, nil
}

func (m *mockSettingsRepo) Set(_ context.Context, key string, value string, _ uuid.UUID) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}

// --- Tests ---

func TestSettingsService_GetAll(t *testing.T) {
	repo := newMockSettingsRepo()
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
	svc := service.NewSettingsService(newMockSettingsRepo(), setupLogger())
	settings, err := svc.GetAll(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(settings) != 0 {
		t.Errorf("expected 0 settings, got %d", len(settings))
	}
}

func TestSettingsService_UpdateMultiple_Success(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := service.NewSettingsService(repo, setupLogger())

	err := svc.UpdateMultiple(context.Background(), map[string]string{
		"theme":    "dark",
		"currency": "USD",
	}, uuid.New())
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

func TestSettingsService_UpdateMultiple_Error(t *testing.T) {
	repo := newMockSettingsRepo()
	repo.setErr = errors.New("db write failed")

	svc := service.NewSettingsService(repo, setupLogger())

	err := svc.UpdateMultiple(context.Background(), map[string]string{
		"theme": "dark",
	}, uuid.New())
	if err == nil {
		t.Error("expected error when Set fails")
	}
}

func TestSettingsService_NilLogger(t *testing.T) {
	// Should not panic with nil logger
	svc := service.NewSettingsService(newMockSettingsRepo(), nil)
	_, err := svc.GetAll(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
