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

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'inv_user', 'hash', 'inv@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Insert a test category so FK constraint is satisfied
	catID := uuid.New()
	_, err = globalPool.Exec(ctx,
		"INSERT INTO categories (id, user_id, name, color) VALUES ($1, $2, 'Software-inv', '#000001') ON CONFLICT DO NOTHING",
		catID, userID)
	if err != nil {
		t.Fatalf("failed to insert test category: %v", err)
	}

	t.Run("Save_and_FindByID", func(t *testing.T) {
		invID := uuid.New()
		inv := entity.Invoice{
			ID:         invID,
			UserID:     userID,
			CategoryID: &catID,
			Vendor:     entity.Vendor{Name: "GitHub"},
			Amount:     15.00,
			Currency:   "USD",
			IssuedAt:   time.Date(2099, 3, 1, 0, 0, 0, 0, time.UTC),
		}

		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("Save: unexpected error: %v", err)
		}

		found, err := repo.FindByID(ctx, invID, userID)
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
		_ = repo.Delete(ctx, invID, userID)
	})

	t.Run("Save_Upsert_UpdatesFields", func(t *testing.T) {
		invID := uuid.New()
		inv := entity.Invoice{
			ID:       invID,
			UserID:   userID,
			Vendor:   entity.Vendor{Name: "InitialVendor"},
			Amount:   10.00,
			Currency: "EUR",
		}
		_ = repo.Save(ctx, inv)

		inv.Amount = 20.00
		if err := repo.Save(ctx, inv); err != nil {
			t.Fatalf("Save upsert: %v", err)
		}
		found, _ := repo.FindByID(ctx, invID, userID)
		if found.Amount != 20.00 {
			t.Errorf("upsert amount: want 20.00, got %f", found.Amount)
		}

		_ = repo.Delete(ctx, invID, userID)
	})

	t.Run("Update_ChangesFields", func(t *testing.T) {
		invID := uuid.New()
		_ = repo.Save(ctx, entity.Invoice{
			ID: invID, UserID: userID, Vendor: entity.Vendor{Name: "OldCo"}, Amount: 5.00, Currency: "EUR",
		})

		err := repo.Update(ctx, entity.Invoice{
			ID:         invID,
			UserID:     userID,
			Vendor:     entity.Vendor{Name: "NewCo"},
			CategoryID: &catID,
			Amount:     99.99,
			Currency:   "EUR",
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		found, _ := repo.FindByID(ctx, invID, userID)
		if found.Vendor.Name != "NewCo" {
			t.Errorf("vendor after update: want 'NewCo', got '%s'", found.Vendor.Name)
		}
		if found.Amount != 99.99 {
			t.Errorf("amount after update: want 99.99, got %f", found.Amount)
		}

		_ = repo.Delete(ctx, invID, userID)
	})

	t.Run("Update_NotFound_ReturnsError", func(t *testing.T) {
		err := repo.Update(ctx, entity.Invoice{ID: uuid.New(), UserID: userID, Currency: "EUR"})
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound, got %v", err)
		}
	})

	t.Run("ExistsByContentHash_FalseWhenAbsent", func(t *testing.T) {
		exists, err := repo.ExistsByContentHash(ctx, "nonexistenthash", userID)
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
			ID: invID, UserID: userID, Vendor: entity.Vendor{Name: "HashCo"},
			Amount: 1, Currency: "EUR", ContentHash: hash,
		})

		exists, err := repo.ExistsByContentHash(ctx, hash, userID)
		if err != nil {
			t.Fatalf("ExistsByContentHash: %v", err)
		}
		if !exists {
			t.Error("expected true after saving with that hash, got false")
		}

		_ = repo.Delete(ctx, invID, userID)
	})

	t.Run("GetOriginalFile_ReturnsFileData", func(t *testing.T) {
		invID := uuid.New()
		fileBytes := []byte("%PDF-fake pdf bytes for test")
		hash := "filehash-" + uuid.New().String()
		_ = repo.Save(ctx, entity.Invoice{
			ID:                  invID,
			UserID:              userID,
			Vendor:              entity.Vendor{Name: "FileCo"},
			Amount:              1,
			Currency:            "EUR",
			ContentHash:         hash,
			OriginalFileName:    "receipt.pdf",
			OriginalFileContent: fileBytes,
		})

		content, mime, name, err := repo.GetOriginalFile(ctx, invID, userID)
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

		_ = repo.Delete(ctx, invID, userID)
	})

	t.Run("GetOriginalFile_NotFound_ReturnsError", func(t *testing.T) {
		_, _, _, err := repo.GetOriginalFile(ctx, uuid.New(), userID)
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound, got %v", err)
		}
	})

	t.Run("FindAll_ReturnsAllSaved", func(t *testing.T) {
		clearTables(ctx, t)
		// Re-create user and category after clear
		_, _ = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'inv_user_2', 'hash', 'inv2@example.com')", userID)
		_, _ = globalPool.Exec(ctx,
			"INSERT INTO categories (id, user_id, name, color) VALUES ($1, $2, 'Software-inv', '#000001') ON CONFLICT DO NOTHING",
			catID, userID)

		for i := 0; i < 3; i++ {
			_ = repo.Save(ctx, entity.Invoice{
				ID:       uuid.New(),
				UserID:   userID,
				Vendor:   entity.Vendor{Name: "Vendor"},
				Amount:   float64(i + 1),
				Currency: "EUR",
			})
		}
		all, err := repo.FindAll(ctx, entity.InvoiceFilter{UserID: userID})
		if err != nil {
			t.Fatalf("FindAll: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("expected 3 invoices, got %d", len(all))
		}
	})

	t.Run("Delete_NotFound_ReturnsError", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New(), userID)
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound, got %v", err)
		}
	})

	t.Run("FindByID_NotFound_ReturnsError", func(t *testing.T) {
		_, err := repo.FindByID(ctx, uuid.New(), userID)
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound, got %v", err)
		}
	})

	t.Run("Sharing_Permissions", func(t *testing.T) {
		clearTables(ctx, t)
		ownerID := uuid.New()
		sharedUserID := uuid.New()
		_, _ = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'owner_inv', 'hash', 'owner_inv@example.com')", ownerID)
		_, _ = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'shared_inv', 'hash', 'shared_inv@example.com')", sharedUserID)

		catID := uuid.New()
		_, _ = globalPool.Exec(ctx, "INSERT INTO categories (id, user_id, name, color) VALUES ($1, $2, 'Shared-inv-cat', '#000001')", catID, ownerID)

		// Create two invoices for the owner
		invDirectID := uuid.New()
		invCatSharedID := uuid.New()

		_ = repo.Save(ctx, entity.Invoice{
			ID:       invDirectID,
			UserID:   ownerID,
			Vendor:   entity.Vendor{Name: "DirectShared"},
			Amount:   100.0,
			Currency: "EUR",
		})
		_ = repo.Save(ctx, entity.Invoice{
			ID:         invCatSharedID,
			UserID:     ownerID,
			CategoryID: &catID,
			Vendor:     entity.Vendor{Name: "CatShared"},
			Amount:     200.0,
			Currency:   "EUR",
		})

		// 1. Direct sharing with 'view' permission
		_, err := globalPool.Exec(ctx, "INSERT INTO shared_invoices (invoice_id, owner_user_id, shared_with_user_id, permission_level) VALUES ($1, $2, $3, 'view')", invDirectID, ownerID, sharedUserID)
		if err != nil {
			t.Fatalf("failed to insert shared_invoices: %v", err)
		}

		// 2. Category sharing with 'edit' permission
		_, err = globalPool.Exec(ctx, "INSERT INTO shared_categories (category_id, owner_user_id, shared_with_user_id, permission_level) VALUES ($1, $2, $3, 'edit')", catID, ownerID, sharedUserID)
		if err != nil {
			t.Fatalf("failed to insert shared_categories: %v", err)
		}

		// 3. User with 'view' access can FindByID
		found, err := repo.FindByID(ctx, invDirectID, sharedUserID)
		if err != nil {
			t.Fatalf("FindByID (shared view): %v", err)
		}
		if found.ID != invDirectID {
			t.Errorf("expected invDirectID")
		}

		// User with 'view' access gets error on Update
		err = repo.Update(ctx, entity.Invoice{
			ID:       invDirectID,
			UserID:   sharedUserID,
			Vendor:   entity.Vendor{Name: "Hacked"},
			Amount:   999.0,
			Currency: "EUR",
		})
		if !errors.Is(err, entity.ErrInvoiceNotFound) {
			t.Errorf("expected ErrInvoiceNotFound for Update with view access, got %v", err)
		}

		// 4. User with 'edit' access can Update
		err = repo.Update(ctx, entity.Invoice{
			ID:         invCatSharedID,
			UserID:     sharedUserID,
			CategoryID: &catID,
			Vendor:     entity.Vendor{Name: "CatSharedEdited"},
			Amount:     250.0,
			Currency:   "EUR",
		})
		if err != nil {
			t.Fatalf("Update (shared edit): %v", err)
		}

		// 5. FindAll with IncludeShared=true
		allShared, err := repo.FindAll(ctx, entity.InvoiceFilter{
			UserID:        sharedUserID,
			IncludeShared: true,
		})
		if err != nil {
			t.Fatalf("FindAll (IncludeShared=true): %v", err)
		}
		if len(allShared) != 2 {
			t.Errorf("expected 2 shared invoices, got %d", len(allShared))
		}

		// 5. FindAll with Source="shared"
		sourceShared, err := repo.FindAll(ctx, entity.InvoiceFilter{
			UserID:        sharedUserID,
			IncludeShared: false,
			Source:        "shared",
		})
		if err != nil {
			t.Fatalf("FindAll (Source=shared): %v", err)
		}
		if len(sourceShared) != 2 {
			t.Errorf("expected 2 source=shared invoices, got %d", len(sourceShared))
		}
	})
}
