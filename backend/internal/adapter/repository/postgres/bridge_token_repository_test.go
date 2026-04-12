package postgres

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestBridgeAccessTokenRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	repo := NewBridgeAccessTokenRepository(globalPool, setupLogger())

	// Create a test user since tokens likely depend on users
	userRepo := NewUserRepository(globalPool, setupLogger())
	userID := uuid.New()
	user := entity.User{
		ID:           userID,
		Username:     "test_token_user",
		PasswordHash: "hash",
		Email:        "token_user@test.local",
		Role:         "user",
	}
	_ = userRepo.Upsert(ctx, user)

	t.Run("Bridge Token Operations", func(t *testing.T) {
		tokenID := uuid.New()
		now := time.Now().Truncate(time.Microsecond)
		token := entity.BridgeAccessToken{
			ID:        tokenID,
			UserID:    userID,
			Name:      "Test Device",
			TokenHash: "test_hash_123",
			CreatedAt: now,
		}

		// 1. Save
		err := repo.Save(ctx, token)
		if err != nil {
			t.Fatalf("Save: expected no error, got %v", err)
		}

		// 2. FindByID
		found, err := repo.FindByID(ctx, tokenID, userID)
		if err != nil {
			t.Fatalf("FindByID: expected no error, got %v", err)
		}
		if found.ID != tokenID {
			t.Errorf("expected ID %v, got %v", tokenID, found.ID)
		}
		if found.Name != "Test Device" {
			t.Errorf("expected Name 'Test Device', got %s", found.Name)
		}

		// 3. FindByHash
		foundHash, err := repo.FindByHash(ctx, "test_hash_123")
		if err != nil {
			t.Fatalf("FindByHash: expected no error, got %v", err)
		}
		if foundHash.ID != tokenID {
			t.Errorf("expected ID %v, got %v", tokenID, foundHash.ID)
		}

		// 4. UpdateLastUsed
		err = repo.UpdateLastUsed(ctx, tokenID, userID)
		if err != nil {
			t.Fatalf("UpdateLastUsed: expected no error, got %v", err)
		}

		foundUpdated, _ := repo.FindByID(ctx, tokenID, userID)
		if foundUpdated.LastUsedAt == nil {
			t.Error("expected LastUsedAt to be updated, but it is nil")
		}

		// 5. FindAll
		otherToken := entity.BridgeAccessToken{
			ID:        uuid.New(),
			UserID:    userID,
			Name:      "Other Device",
			TokenHash: "other_hash",
			CreatedAt: time.Now(),
		}
		_ = repo.Save(ctx, otherToken)

		allTokens, err := repo.FindAll(ctx, userID)
		if err != nil {
			t.Fatalf("FindAll: expected no error, got %v", err)
		}
		if len(allTokens) != 2 {
			t.Errorf("expected 2 tokens, got %d", len(allTokens))
		}

		// 6. Delete
		err = repo.Delete(ctx, tokenID, userID)
		if err != nil {
			t.Fatalf("Delete: expected no error, got %v", err)
		}

		_, err = repo.FindByID(ctx, tokenID, userID)
		if err == nil {
			t.Error("FindByID: expected error after deletion, got nil")
		}
	})
}
