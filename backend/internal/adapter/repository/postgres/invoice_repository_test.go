package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestInvoiceRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	repo := NewInvoiceRepository(globalPool, setupLogger())

	// Insert a test category so FK constraint is satisfied
	catID := uuid.New()
	_, err := globalPool.Exec(ctx,
		"INSERT INTO categories (id, name, color) VALUES ($1, 'Software-inv', '#000001') ON CONFLICT DO NOTHING",
		catID)
	if err != nil {
		t.Fatalf("failed to insert test category: %v", err)
	}

	t.Run("Save_and_FindByID", func(t *testing.T) {
		invID := uuid.New()
		inv := entity.Invoice{
			ID:         invID,
			RawText:    "Invoice data 123",
			CategoryID: &catID,
			Vendor:     entity.Vendor{Name: "GitHub"},
			Amount:     15.00,
			Currency:   "USD",
			IssuedAt:   time.Date(2099, 3, 1, 0, 0, 0, 0, time.UTC),
		}

		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("Save: unexpected error: %v", err)
		}

		found, err := repo.FindByID(ctx, invID)
		if err != nil {
			t.Fatalf("FindByID: unexpected error: %v", err)
		}
		if found.Vendor.Name != "GitHub" {
			t.Errorf("vendor: want 'GitHub', got '%s'", found.Vendor.Name)
		}
		if found.CategoryID == nil || *found.CategoryID != catID {
			t.Errorf("categoryID: want %v, got %v", catID, found.CategoryID)
		}
		if found.Amount != 15.00 {
			t.Errorf("amount: want 15.00, got %f", found.Amount)
		}

		// cleanup
		_ = repo.Delete(ctx, invID)
	})

	t.Run("Save_Upsert_UpdatesFields", func(t *testing.T) {
		invID := uuid.New()
		inv := entity.Invoice{
			ID:       invID,
			RawText:  "initial text",
			Vendor:   entity.Vendor{Name: "InitialVendor"},
			Amount:   10.00,
			Currency: "EUR",
		}
		_ = repo.Save(ctx, inv)

		inv.Amount = 20.00
		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("Save upsert: %v", err)
		}
		found, _ := repo.FindByID(ctx, invID)
		if found.Amount != 20.00 {
			t.Errorf("upsert amount: want 20.00, got %f", found.Amount)
		}

		_ = repo.Delete(ctx, invID)
	})

	t.Run("Update_ChangesFields", func(t *testing.T) {
		invID := uuid.New()
		_ = repo.Save(ctx, entity.Invoice{
			ID: invID, Vendor: entity.Vendor{Name: "OldCo"}, Amount: 5.00, Currency: "EUR",
		})

		err := repo.Update(ctx, entity.Invoice{
			ID:         invID,
			Vendor:     entity.Vendor{Name: "NewCo"},
			CategoryID: &catID,
			Amount:     99.99,
			Currency:   "EUR",
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		found, _ := repo.FindByID(ctx, invID)
		if found.Vendor.Name != "NewCo" {
			t.Errorf("vendor after update: want 'NewCo', got '%s'", found.Vendor.Name)
		}
		if found.Amount != 99.99 {
			t.Errorf("amount after update: want 99.99, got %f", found.Amount)
		}

		_ = repo.Delete(ctx, invID)
	})

	t.Run("Update_NotFound_ReturnsError", func(t *testing.T) {
		err := repo.Update(ctx, entity.Invoice{ID: uuid.New(), Currency: "EUR"})
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound, got %v", err)
		}
	})

	t.Run("ExistsByContentHash_FalseWhenAbsent", func(t *testing.T) {
		exists, err := repo.ExistsByContentHash(ctx, "nonexistenthash")
		if err != nil {
			t.Fatalf("ExistsByContentHash: %v", err)
		}
		if exists {
			t.Error("expected false for absent hash, got true")
		}
	})

	t.Run("ExistsByContentHash_TrueAfterSave", func(t *testing.T) {
		hash := "abc123uniquehash-" + uuid.New().String()
		invID := uuid.New()
		_ = repo.Save(ctx, entity.Invoice{
			ID: invID, Vendor: entity.Vendor{Name: "HashCo"},
			Amount: 1, Currency: "EUR", ContentHash: hash,
		})

		exists, err := repo.ExistsByContentHash(ctx, hash)
		if err != nil {
			t.Fatalf("ExistsByContentHash: %v", err)
		}
		if !exists {
			t.Error("expected true after saving with that hash, got false")
		}

		_ = repo.Delete(ctx, invID)
	})

	t.Run("GetOriginalFile_ReturnsFileData", func(t *testing.T) {
		invID := uuid.New()
		fileBytes := []byte("fake pdf bytes for test")
		hash := "filehash-" + uuid.New().String()
		_ = repo.Save(ctx, entity.Invoice{
			ID:                  invID,
			Vendor:              entity.Vendor{Name: "FileCo"},
			Amount:              1,
			Currency:            "EUR",
			ContentHash:         hash,
			OriginalFileName:    "receipt.pdf",
			OriginalFileMime:    "application/pdf",
			OriginalFileSize:    int64(len(fileBytes)),
			OriginalFileContent: fileBytes,
		})

		content, mime, name, err := repo.GetOriginalFile(ctx, invID)
		if err != nil {
			t.Fatalf("GetOriginalFile: %v", err)
		}
		if string(content) != string(fileBytes) {
			t.Error("file content mismatch")
		}
		if mime != "application/pdf" {
			t.Errorf("mime: want 'application/pdf', got '%s'", mime)
		}
		if name != "receipt.pdf" {
			t.Errorf("name: want 'receipt.pdf', got '%s'", name)
		}

		_ = repo.Delete(ctx, invID)
	})

	t.Run("GetOriginalFile_NotFound_ReturnsError", func(t *testing.T) {
		_, _, _, err := repo.GetOriginalFile(ctx, uuid.New())
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound, got %v", err)
		}
	})

	t.Run("FindAll_ReturnsAllSaved", func(t *testing.T) {
		clearTables(ctx, t)
		// Re-create category after clear
		_, _ = globalPool.Exec(ctx,
			"INSERT INTO categories (id, name, color) VALUES ($1, 'Software-inv', '#000001') ON CONFLICT DO NOTHING",
			catID)

		for i := 0; i < 3; i++ {
			_ = repo.Save(ctx, entity.Invoice{
				ID:       uuid.New(),
				Vendor:   entity.Vendor{Name: "Vendor"},
				Amount:   float64(i + 1),
				Currency: "EUR",
			})
		}
		all, err := repo.FindAll(ctx)
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("expected 3 invoices, got %d", len(all))
		}
	})

	t.Run("Delete_NotFound_ReturnsError", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New())
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound, got %v", err)
		}
	})

	t.Run("FindByID_NotFound_ReturnsError", func(t *testing.T) {
		_, err := repo.FindByID(ctx, uuid.New())
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound, got %v", err)
		}
	})
}
