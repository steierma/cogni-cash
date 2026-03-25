package postgres

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestUserRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	repo := NewUserRepository(globalPool, setupLogger())

	t.Run("Upsert and Queries", func(t *testing.T) {
		userID := uuid.New()
		user := entity.User{
			ID:           userID,
			Username:     "admin",
			PasswordHash: "initial_hash",
			Email:        "admin@test.local",
			FullName:     "Test Admin",
			Address:      "123 Localhost Ave",
			Role:         "admin",
		}

		// 1. Upsert (Insert)
		err := repo.Upsert(ctx, user)
		if err != nil {
			t.Fatalf("expected no error upserting user, got: %v", err)
		}

		// 2. FindByUsername
		found, err := repo.FindByUsername(ctx, "admin")
		if err != nil {
			t.Fatalf("expected no error finding user, got: %v", err)
		}
		if found.PasswordHash != "initial_hash" {
			t.Errorf("expected initial_hash, got %s", found.PasswordHash)
		}
		if found.Email != "admin@test.local" {
			t.Errorf("expected admin@test.local, got %s", found.Email)
		}

		// 3. Upsert (Update on conflict)
		// password_hash must NOT be overwritten — only profile fields are updated.
		user.PasswordHash = "should_not_be_stored"
		user.FullName = "Updated Admin"
		err = repo.Upsert(ctx, user)
		if err != nil {
			t.Fatalf("expected no error upserting user, got: %v", err)
		}

		foundByID, _ := repo.FindByID(ctx, found.ID)
		// Password must remain the original value set on insert.
		if foundByID.PasswordHash != "initial_hash" {
			t.Errorf("Upsert must not overwrite password_hash: expected initial_hash, got %s", foundByID.PasswordHash)
		}
		// Profile fields are still updated on conflict.
		if foundByID.FullName != "Updated Admin" {
			t.Errorf("expected Updated Admin, got %s", foundByID.FullName)
		}

		// 4. UpdatePassword
		err = repo.UpdatePassword(ctx, found.ID, "final_hash")
		if err != nil {
			t.Fatalf("expected no error updating password, got: %v", err)
		}

		finalUser, _ := repo.FindByID(ctx, found.ID)
		if finalUser.PasswordHash != "final_hash" {
			t.Errorf("expected final_hash, got %s", finalUser.PasswordHash)
		}
	})
}
