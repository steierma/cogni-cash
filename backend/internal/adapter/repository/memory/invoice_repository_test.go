package memory

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestInvoiceRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewInvoiceRepository()
	userID := uuid.New()

	t.Run("Save_and_FindByID", func(t *testing.T) {
		inv := entity.Invoice{
			ID:       uuid.New(),
			UserID:   userID,
			Vendor:   entity.Vendor{Name: "Test Vendor"},
			Amount:   100.0,
			Currency: "EUR",
			IssuedAt: time.Now(),
		}

		err := repo.Save(ctx, inv)
		if err != nil {
			t.Fatalf("Save: %v", err)
		}

		found, err := repo.FindByID(ctx, inv.ID, userID)
		if err != nil {
			t.Fatalf("FindByID: %v", err)
		}
		if found.Vendor.Name != "Test Vendor" {
			t.Errorf("expected 'Test Vendor', got %s", found.Vendor.Name)
		}
	})

	t.Run("Isolation", func(t *testing.T) {
		inv := entity.Invoice{ID: uuid.New(), UserID: userID, Vendor: entity.Vendor{Name: "Private"}}
		_ = repo.Save(ctx, inv)

		_, err := repo.FindByID(ctx, inv.ID, uuid.New())
		if err == nil {
			t.Error("FindByID: expected error for other user")
		}
	})

	t.Run("Update", func(t *testing.T) {
		inv := entity.Invoice{ID: uuid.New(), UserID: userID, Vendor: entity.Vendor{Name: "Old"}}
		_ = repo.Save(ctx, inv)

		inv.Vendor.Name = "New"
		err := repo.Update(ctx, inv)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		found, _ := repo.FindByID(ctx, inv.ID, userID)
		if found.Vendor.Name != "New" {
			t.Errorf("expected 'New', got %s", found.Vendor.Name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		inv := entity.Invoice{ID: uuid.New(), UserID: userID}
		_ = repo.Save(ctx, inv)

		err := repo.Delete(ctx, inv.ID, userID)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		_, err = repo.FindByID(ctx, inv.ID, userID)
		if err == nil {
			t.Error("FindByID: expected error after deletion")
		}
	})
}
