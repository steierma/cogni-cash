package postgres

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestInvoiceRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	repo := NewInvoiceRepository(globalPool, setupLogger())

	// Insert a test category to ensure we have a valid UUID for the foreign key constraint
	catID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO categories (id, name, color) VALUES ($1, 'Software', '#000000') ON CONFLICT DO NOTHING", catID)
	if err != nil {
		t.Fatalf("failed to insert test category: %v", err)
	}

	t.Run("Save, Find, and Delete Invoice", func(t *testing.T) {
		invID := uuid.New()
		inv := entity.Invoice{
			ID:         invID,
			RawText:    "Invoice data 123",
			CategoryID: &catID, // Updated to use the UUID pointer
			Vendor:     entity.Vendor{Name: "GitHub"},
			Amount:     15.00,
			Currency:   "USD",
			IssuedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		}

		// 1. Save
		err := repo.Save(ctx, inv)
		if err != nil {
			t.Fatalf("expected no error saving invoice, got: %v", err)
		}

		// 2. FindByID
		found, err := repo.FindByID(ctx, invID)
		if err != nil {
			t.Fatalf("expected no error finding invoice, got: %v", err)
		}
		if found.Vendor.Name != "GitHub" {
			t.Errorf("expected vendor GitHub, got %s", found.Vendor.Name)
		}
		if found.CategoryID == nil || *found.CategoryID != catID {
			t.Errorf("expected CategoryID %v, got %v", catID, found.CategoryID)
		}

		// 3. Upsert (Save again to trigger ON CONFLICT DO UPDATE)
		inv.Amount = 20.00
		err = repo.Save(ctx, inv)
		if err != nil {
			t.Fatalf("expected no error updating invoice, got: %v", err)
		}

		foundUpdated, _ := repo.FindByID(ctx, invID)
		if foundUpdated.Amount != 20.00 {
			t.Errorf("expected updated amount 20.00, got %f", foundUpdated.Amount)
		}

		// 4. FindAll
		all, err := repo.FindAll(ctx)
		if err != nil {
			t.Fatalf("expected no error finding all invoices, got: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 invoice, got %d", len(all))
		}

		// 5. Delete
		err = repo.Delete(ctx, invID)
		if err != nil {
			t.Fatalf("expected no error deleting invoice, got: %v", err)
		}

		_, err = repo.FindByID(ctx, invID)
		if err == nil {
			t.Error("expected error finding deleted invoice, got nil")
		}
	})
}
