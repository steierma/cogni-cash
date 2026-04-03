package postgres

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestPayslipRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	repo := NewPayslipRepository(globalPool)

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'payslip_user', 'hash', 'ps@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	t.Run("Save, Find, Update, Delete Lifecycle", func(t *testing.T) {
		// 1. Setup dummy entity matching our real-world data
		payslip := entity.Payslip{
			UserID:              userID,
			OriginalFileName:    "99999999_Monatsabrechnung_202602.pdf",
			OriginalFileContent: []byte("dummy pdf binary content"),
			ContentHash:         "dummy_sha256_hash_12345",
			PeriodMonthNum:      2,
			PeriodYear:          2026,
			EmployerName:        "Test Employer",
			TaxClass:            "4",
			TaxID:               "41935678240",
			GrossPay:            8812.58,
			NetPay:              5770.44,
			PayoutAmount:        4752.01,
			CustomDeductions:    -438.94,
			Bonuses: []entity.Bonus{
				{Description: "Einmalzahlung", Amount: 600.00},
			},
		}

		// 2. Save
		err := repo.Save(ctx, &payslip)
		if err != nil {
			t.Fatalf("expected no error saving payslip, got: %v", err)
		}
		if payslip.ID == "" {
			t.Fatalf("expected populated UUID after save")
		}

		// 3. ExistsByHash
		exists, err := repo.ExistsByHash(ctx, "dummy_sha256_hash_12345", userID)
		if err != nil || !exists {
			t.Errorf("expected hash to exist in database")
		}

		// 4. FindByID — also verify Bonuses are loaded
		found, err := repo.FindByID(ctx, payslip.ID, userID)
		if err != nil {
			t.Fatalf("expected no error finding by ID, got: %v", err)
		}
		if found.NetPay != 5770.44 {
			t.Errorf("expected NetPay 5770.44, got %f", found.NetPay)
		}
		if len(found.Bonuses) != 1 {
			t.Errorf("expected 1 Bonus after save, got %d", len(found.Bonuses))
		} else {
			if found.Bonuses[0].Description != "Einmalzahlung" {
				t.Errorf("expected description 'Einmalzahlung', got '%s'", found.Bonuses[0].Description)
			}
			if found.Bonuses[0].Amount != 600.00 {
				t.Errorf("expected amount 600.00, got %f", found.Bonuses[0].Amount)
			}
		}

		// 5. Update — replace Bonuses with TZUG entries and change TaxClass
		found.TaxClass = "3"
		found.Bonuses = []entity.Bonus{
			{Description: "TZUG Zusatzbetrag", Amount: 398.50},
			{Description: "TZUG-Tarifl. Zusatzgeld", Amount: 1581.60},
			{Description: "Urlaubsgeld", Amount: 793.38},
		}
		err = repo.Update(ctx, &found)
		if err != nil {
			t.Fatalf("expected no error updating payslip, got: %v", err)
		}
		updated, _ := repo.FindByID(ctx, payslip.ID, userID)
		if updated.TaxClass != "3" {
			t.Errorf("expected updated TaxClass 3, got %s", updated.TaxClass)
		}
		if len(updated.Bonuses) != 3 {
			t.Errorf("expected 3 Bonuses after update, got %d", len(updated.Bonuses))
		} else {
			expectedBonuses := []struct {
				desc   string
				amount float64
			}{
				{"TZUG Zusatzbetrag", 398.50},
				{"TZUG-Tarifl. Zusatzgeld", 1581.60},
				{"Urlaubsgeld", 793.38},
			}
			for i, exp := range expectedBonuses {
				if updated.Bonuses[i].Description != exp.desc {
					t.Errorf("Bonus[%d]: expected description '%s', got '%s'", i, exp.desc, updated.Bonuses[i].Description)
				}
				if updated.Bonuses[i].Amount != exp.amount {
					t.Errorf("Bonus[%d]: expected amount %.2f, got %.2f", i, exp.amount, updated.Bonuses[i].Amount)
				}
			}
		}

		// 5b. Update — clearing all Bonuses should work too
		updated.Bonuses = []entity.Bonus{}
		if err = repo.Update(ctx, &updated); err != nil {
			t.Fatalf("expected no error clearing bonuses, got: %v", err)
		}
		afterClear, _ := repo.FindByID(ctx, payslip.ID, userID)
		if len(afterClear.Bonuses) != 0 {
			t.Errorf("expected 0 Bonuses after clear, got %d", len(afterClear.Bonuses))
		}

		// 6. GetOriginalFile
		content, mime, filename, err := repo.GetOriginalFile(ctx, payslip.ID, userID)
		if err != nil {
			t.Fatalf("expected no error fetching file, got: %v", err)
		}
		if string(content) != "dummy pdf binary content" || mime != "application/pdf" || filename != "99999999_Monatsabrechnung_202602.pdf" {
			t.Errorf("file metadata mismatch during retrieval")
		}

		// 7. FindAll
		all, err := repo.FindAll(ctx, entity.PayslipFilter{UserID: userID})
		if err != nil || len(all) == 0 {
			t.Fatalf("expected at least 1 payslip in FindAll, got %d", len(all))
		}

		// 7b. FindAll with filtering
		filtered, err := repo.FindAll(ctx, entity.PayslipFilter{UserID: userID, Employer: "NonExistent"})
		if err != nil || len(filtered) != 0 {
			t.Errorf("expected 0 results for non-existent employer, got %d", len(filtered))
		}

		// 8. Delete (should cascade and remove Bonuses as well)
		err = repo.Delete(ctx, payslip.ID, userID)
		if err != nil {
			t.Fatalf("expected no error deleting payslip, got: %v", err)
		}
		_, err = repo.FindByID(ctx, payslip.ID, userID)
		if err == nil {
			t.Errorf("expected error when querying deleted payslip, got nil")
		}
	})
}
