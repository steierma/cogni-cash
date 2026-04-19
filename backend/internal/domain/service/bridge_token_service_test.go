package service_test

import (
	"context"
	"testing"

	"cogni-cash/internal/adapter/repository/memory"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

func TestBridgeAccessTokenService(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := memory.NewBridgeAccessTokenRepository()
	svc := service.NewBridgeAccessTokenService(repo)

	t.Run("CreateToken", func(t *testing.T) {
		resp, err := svc.CreateToken(ctx, userID, "My Test Device")
		if err != nil {
			t.Fatalf("CreateToken: %v", err)
		}

		if resp.Token == "" {
			t.Error("CreateToken: token is empty")
		}

		if resp.Info.Name != "My Test Device" {
			t.Errorf("CreateToken: expected Name 'My Test Device', got %s", resp.Info.Name)
		}
		if resp.Info.UserID != userID {
			t.Errorf("CreateToken: wrong user ID")
		}
		if resp.Info.TokenHash == "" {
			t.Error("CreateToken: hash is empty")
		}

		tokens, err := svc.ListTokens(ctx, userID)
		if err != nil {
			t.Fatalf("ListTokens: %v", err)
		}
		if len(tokens) != 1 {
			t.Errorf("ListTokens: expected 1, got %d", len(tokens))
		}
	})

	t.Run("ValidateToken", func(t *testing.T) {
		resp, _ := svc.CreateToken(ctx, userID, "Device 2")

		validUserID, err := svc.ValidateToken(ctx, resp.Token)
		if err != nil {
			t.Fatalf("ValidateToken: %v", err)
		}
		if validUserID != userID {
			t.Errorf("ValidateToken: expected userID %v, got %v", userID, validUserID)
		}

		_, err = svc.ValidateToken(ctx, "invalid-token-string")
		if err == nil {
			t.Error("ValidateToken: expected error for invalid token")
		}
	})

	t.Run("RevokeToken", func(t *testing.T) {
		resp, _ := svc.CreateToken(ctx, userID, "Device 3")

		err := svc.RevokeToken(ctx, resp.Info.ID, userID)
		if err != nil {
			t.Fatalf("RevokeToken: %v", err)
		}

		_, err = svc.ValidateToken(ctx, resp.Token)
		if err == nil {
			t.Error("ValidateToken: expected error after revoking token")
		}
	})
}
