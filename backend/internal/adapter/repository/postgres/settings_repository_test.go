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
	repo := NewSettingsRepository(globalPool, userRepo, setupLogger())

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
	err := repo.Set(ctx, "smtp_host", "admin-host", adminID)
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
	err = repo.Set(ctx, "smtp_host", "user-host", userID)
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
