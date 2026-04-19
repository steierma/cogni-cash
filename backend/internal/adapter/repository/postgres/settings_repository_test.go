package postgres

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestSettingsRepository_Fallback(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	userRepo := NewUserRepository(globalPool, setupLogger())
	repo := NewSettingsRepository(globalPool, userRepo, "test-vault-key", setupLogger())

	// 1. Setup Admin
	adminID := uuid.New()
	admin := entity.User{
		ID:       adminID,
		Username: "admin",
		Email:    "admin@test.local",
	}
	_ = userRepo.Upsert(ctx, admin)

	// 2. Setup regular user
	userID := uuid.New()
	user := entity.User{
		ID:       userID,
		Username: "user",
		Email:    "user@test.local",
	}
	_ = userRepo.Upsert(ctx, user)

	// 3. Set global (admin) setting
	err := repo.Set(ctx, "smtp_host", "admin-host", adminID, false)
	if err != nil {
		t.Fatalf("failed to set admin setting: %v", err)
	}

	// 4. Get setting for user (should fallback to admin)
	val, err := repo.Get(ctx, "smtp_host", userID)
	if err != nil {
		t.Fatalf("failed to get setting with fallback: %v", err)
	}
	if val != "admin-host" {
		t.Errorf("expected fallback value 'admin-host', got '%s'", val)
	}

	// 5. Set user-specific override
	err = repo.Set(ctx, "smtp_host", "user-host", userID, false)
	if err != nil {
		t.Fatalf("failed to set user setting: %v", err)
	}

	// 6. Get setting for user (should NOT fallback)
	val, err = repo.Get(ctx, "smtp_host", userID)
	if err != nil {
		t.Fatalf("failed to get user setting: %v", err)
	}
	if val != "user-host" {
		t.Errorf("expected user value 'user-host', got '%s'", val)
	}
}

func TestSettingsRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	userRepo := NewUserRepository(globalPool, setupLogger())
	repo := NewSettingsRepository(globalPool, userRepo, "test-vault-key", setupLogger())

	userID := uuid.New()
	_ = userRepo.Upsert(ctx, entity.User{ID: userID, Username: "user", Email: "u@test.local"})

	_ = repo.Set(ctx, "theme", "dark", userID, false)
	_ = repo.Set(ctx, "smtp_pass", "secret", userID, true)

	settings, err := repo.GetAll(ctx, userID)
	if err != nil {
		t.Fatalf("failed to GetAll: %v", err)
	}

	if len(settings) != 2 {
		t.Errorf("expected 2 settings, got %d", len(settings))
	}
	if settings["theme"] != "dark" {
		t.Errorf("expected theme dark, got %s", settings["theme"])
	}
	if settings["smtp_pass"] != "secret" {
		t.Errorf("expected decrypted smtp_pass secret, got %s", settings["smtp_pass"])
	}
}

func TestSettingsRepository_EncryptionDecryption(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	userRepo := NewUserRepository(globalPool, setupLogger())
	repo := NewSettingsRepository(globalPool, userRepo, "test-vault-key", setupLogger())

	userID := uuid.New()
	_ = userRepo.Upsert(ctx, entity.User{ID: userID, Username: "user", Email: "u@test.local"})

	secretValue := "super-secret-password"
	err := repo.Set(ctx, "smtp_password", secretValue, userID, true)
	if err != nil {
		t.Fatalf("failed to set sensitive setting: %v", err)
	}

	// Verify encryption at rest (query the DB directly)
	var rawValue []byte
	err = globalPool.QueryRow(ctx, "SELECT value FROM settings WHERE key = 'smtp_password' AND user_id = $1", userID).Scan(&rawValue)
	if err != nil {
		t.Fatalf("failed to query raw value: %v", err)
	}

	if string(rawValue) == secretValue {
		t.Error("expected value to be encrypted in database, but it was stored as plain text")
	}

	// Verify transparent decryption via Get
	val, err := repo.Get(ctx, "smtp_password", userID)
	if err != nil {
		t.Fatalf("failed to get decrypted setting: %v", err)
	}
	if val != secretValue {
		t.Errorf("expected decrypted value '%s', got '%s'", secretValue, val)
	}
}

func TestSettingsRepository_LegacyFallback(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	userRepo := NewUserRepository(globalPool, setupLogger())
	repo := NewSettingsRepository(globalPool, userRepo, "test-vault-key", setupLogger())

	userID := uuid.New()
	_ = userRepo.Upsert(ctx, entity.User{ID: userID, Username: "user", Email: "u@test.local"})

	// Manually insert a record that is marked sensitive but contains plain text (not encrypted)
	// This simulates legacy data before encryption was enforced.
	_, err := globalPool.Exec(ctx, "INSERT INTO settings (key, value, user_id, is_sensitive) VALUES ($1, $2, $3, $4)",
		"legacy_key", []byte("plain-text-secret"), userID, true)
	if err != nil {
		t.Fatalf("failed to insert legacy record: %v", err)
	}

	// Get should fallback to raw value if decryption fails
	val, err := repo.Get(ctx, "legacy_key", userID)
	if err != nil {
		t.Fatalf("expected fallback to raw value, got error: %v", err)
	}
	if val != "plain-text-secret" {
		t.Errorf("expected legacy value 'plain-text-secret', got '%s'", val)
	}

	// GetAll should also fallback
	settings, err := repo.GetAll(ctx, userID)
	if err != nil {
		t.Fatalf("expected GetAll fallback, got error: %v", err)
	}
	if settings["legacy_key"] != "plain-text-secret" {
		t.Errorf("expected legacy GetAll value 'plain-text-secret', got '%s'", settings["legacy_key"])
	}
}

func TestSettingsRepository_Get_AdminSelf(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	userRepo := NewUserRepository(globalPool, setupLogger())
	repo := NewSettingsRepository(globalPool, userRepo, "test-vault-key", setupLogger())

	adminID := uuid.New()
	_ = userRepo.Upsert(ctx, entity.User{ID: adminID, Username: "admin", Email: "a@test.local"})

	_ = repo.Set(ctx, "theme", "dark", adminID, false)

	val, err := repo.Get(ctx, "theme", adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "dark" {
		t.Errorf("expected dark, got %s", val)
	}
}

func TestSettingsRepository_GetAll_NoUser(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	userRepo := NewUserRepository(globalPool, setupLogger())
	repo := NewSettingsRepository(globalPool, userRepo, "test-vault-key", setupLogger())

	settings, err := repo.GetAll(ctx, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(settings) != 0 {
		t.Errorf("expected 0 settings, got %d", len(settings))
	}
}

func TestSettingsRepository_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	userRepo := NewUserRepository(globalPool, setupLogger())
	repo := NewSettingsRepository(globalPool, userRepo, "test-vault-key", setupLogger())

	userID := uuid.New()
	_ = userRepo.Upsert(ctx, entity.User{ID: userID, Username: "user", Email: "u@test.local"})

	val, err := repo.Get(ctx, "non-existent", userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string, got %s", val)
	}
}
