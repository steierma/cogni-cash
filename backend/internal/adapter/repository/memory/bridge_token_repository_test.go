package memory

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestBridgeAccessTokenRepository(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()

	t.Run("Save and FindByID", func(t *testing.T) {
		repo := NewBridgeAccessTokenRepository()
		token := entity.BridgeAccessToken{
			ID:        uuid.New(),
			UserID:    userID,
			Name:      "My Device",
			TokenHash: "testhash123",
			CreatedAt: time.Now(),
		}

		err := repo.Save(ctx, token)
		if err != nil {
			t.Fatalf("Save: %v", err)
		}

		found, err := repo.FindByID(ctx, token.ID, userID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found.ID != token.ID || found.Name != "My Device" {
			t.Errorf("Unexpected token found: %+v", found)
		}

		_, err = repo.FindByID(ctx, token.ID, otherUserID)
		if err == nil {
			t.Error("FindByID: expected error for wrong user ID")
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		repo := NewBridgeAccessTokenRepository()
		repo.Save(ctx, entity.BridgeAccessToken{ID: uuid.New(), UserID: userID, Name: "D1"})
		repo.Save(ctx, entity.BridgeAccessToken{ID: uuid.New(), UserID: userID, Name: "D2"})
		repo.Save(ctx, entity.BridgeAccessToken{ID: uuid.New(), UserID: otherUserID, Name: "D3"})

		tokens, err := repo.FindAll(ctx, userID)
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		if len(tokens) != 2 {
			t.Errorf("expected 2 tokens, got %d", len(tokens))
		}
	})

	t.Run("FindByHash", func(t *testing.T) {
		repo := NewBridgeAccessTokenRepository()
		token := entity.BridgeAccessToken{ID: uuid.New(), UserID: userID, TokenHash: "myhash"}
		repo.Save(ctx, token)

		found, err := repo.FindByHash(ctx, "myhash")
		if err != nil {
			t.Fatalf("FindByHash: %v", err)
		}
		if found.ID != token.ID {
			t.Errorf("expected ID %v, got %v", token.ID, found.ID)
		}

		_, err = repo.FindByHash(ctx, "invalid")
		if err == nil {
			t.Error("FindByHash: expected error for invalid hash")
		}
	})

	t.Run("UpdateLastUsed", func(t *testing.T) {
		repo := NewBridgeAccessTokenRepository()
		token := entity.BridgeAccessToken{ID: uuid.New(), UserID: userID}
		repo.Save(ctx, token)

		err := repo.UpdateLastUsed(ctx, token.ID, userID)
		if err != nil {
			t.Fatalf("UpdateLastUsed: %v", err)
		}

		found, _ := repo.FindByID(ctx, token.ID, userID)
		if found.LastUsedAt == nil {
			t.Error("LastUsedAt not updated")
		}

		err = repo.UpdateLastUsed(ctx, token.ID, otherUserID)
		if err == nil {
			t.Error("UpdateLastUsed: expected error for wrong user ID")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		repo := NewBridgeAccessTokenRepository()
		token := entity.BridgeAccessToken{ID: uuid.New(), UserID: userID}
		repo.Save(ctx, token)

		err := repo.Delete(ctx, token.ID, otherUserID)
		if err == nil {
			t.Error("Delete: expected error for wrong user ID")
		}

		err = repo.Delete(ctx, token.ID, userID)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		_, err = repo.FindByID(ctx, token.ID, userID)
		if err == nil {
			t.Error("FindByID: expected error after deletion")
		}
	})
}
