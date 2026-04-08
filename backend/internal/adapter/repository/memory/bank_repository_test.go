package memory

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestBankRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewBankRepository()
	userID := uuid.New()

	t.Run("Create_and_GetConnection", func(t *testing.T) {
		conn := &entity.BankConnection{
			ID:             uuid.New(),
			UserID:         userID,
			InstitutionID:  "bank1",
			RequisitionID: "req1",
		}
		err := repo.CreateConnection(ctx, conn)
		if err != nil {
			t.Fatalf("CreateConnection: %v", err)
		}

		found, err := repo.GetConnection(ctx, conn.ID, userID)
		if err != nil {
			t.Fatalf("GetConnection: %v", err)
		}
		if found.InstitutionID != "bank1" {
			t.Errorf("InstitutionID: want bank1, got %s", found.InstitutionID)
		}
	})

	t.Run("Upsert_and_GetAccount", func(t *testing.T) {
		connID := uuid.New()
		_ = repo.CreateConnection(ctx, &entity.BankConnection{ID: connID, UserID: userID})

		acc := entity.BankAccount{
			ID:                uuid.New(),
			ConnectionID:      connID,
			ProviderAccountID: "acc1",
			IBAN:              "DE123",
		}
		err := repo.UpsertAccounts(ctx, []entity.BankAccount{acc})
		if err != nil {
			t.Fatalf("UpsertAccounts: %v", err)
		}

		found, err := repo.GetAccountByID(ctx, acc.ID, userID)
		if err != nil {
			t.Fatalf("GetAccountByID: %v", err)
		}
		if found.IBAN != "DE123" {
			t.Errorf("IBAN: want DE123, got %s", found.IBAN)
		}
	})

	t.Run("DeleteConnection_CleansAccounts", func(t *testing.T) {
		connID := uuid.New()
		_ = repo.CreateConnection(ctx, &entity.BankConnection{ID: connID, UserID: userID})
		accID := uuid.New()
		_ = repo.UpsertAccounts(ctx, []entity.BankAccount{{ID: accID, ConnectionID: connID, ProviderAccountID: "p1"}})

		err := repo.DeleteConnection(ctx, connID, userID)
		if err != nil {
			t.Fatalf("DeleteConnection: %v", err)
		}

		_, err = repo.GetConnection(ctx, connID, userID)
		if err == nil {
			t.Error("GetConnection: expected error after delete")
		}

		_, err = repo.GetAccountByID(ctx, accID, userID)
		if err == nil {
			t.Error("GetAccountByID: expected error after connection delete")
		}
	})
}
