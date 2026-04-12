package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"cogni-cash/internal/domain/entity"
)

func TestPlannedTransactionRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	repo := NewPlannedTransactionRepository(globalPool, setupLogger())

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'pt_user', 'hash', 'pt@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	otherUserID := uuid.New()
	_, err = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'other_pt_user', 'hash', 'other_pt@example.com')", otherUserID)
	if err != nil {
		t.Fatalf("failed to insert other user: %v", err)
	}

	t.Run("Create_and_GetByID", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			ID:          uuid.New(),
			UserID:      userID,
			Amount:      150.50,
			Date:        time.Date(2099, 12, 1, 0, 0, 0, 0, time.UTC),
			Description: "Future expense",
			Status:      entity.PlannedTransactionStatusPending,
		}

		err := repo.Create(ctx, pt)
		if err != nil {
			t.Fatalf("Create: unexpected error: %v", err)
		}
		if pt.CreatedAt.IsZero() {
			t.Errorf("CreatedAt was not set")
		}

		found, err := repo.GetByID(ctx, pt.ID, userID)
		if err != nil {
			t.Fatalf("GetByID: unexpected error: %v", err)
		}
		if found.Amount != 150.50 {
			t.Errorf("amount: want 150.50, got %f", found.Amount)
		}
	})

	t.Run("Isolation", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			ID:          uuid.New(),
			UserID:      userID,
			Amount:      100.0,
			Date:        time.Date(2099, 11, 1, 0, 0, 0, 0, time.UTC),
			Description: "User's Private PT",
		}
		_ = repo.Create(ctx, pt)

		// Try to find as other user
		_, err := repo.GetByID(ctx, pt.ID, otherUserID)
		if err == nil {
			t.Error("GetByID: expected error for other user, got nil")
		}

		// Try to update as other user
		err = repo.Update(ctx, &entity.PlannedTransaction{
			ID:          pt.ID,
			UserID:      otherUserID,
			Amount:      200.0,
			Date:        time.Now(),
			Description: "Hijack",
		})
		if err == nil {
			t.Error("Update: expected error when updating other user's PT, got nil")
		}

		// Try to delete as other user
		err = repo.Delete(ctx, pt.ID, otherUserID)
		if err == nil {
			t.Error("Delete: expected error when deleting other user's PT, got nil")
		}
	})

	t.Run("Update", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			ID:          uuid.New(),
			UserID:      userID,
			Amount:      50.0,
			Date:        time.Now().UTC(),
			Description: "To Update",
		}
		_ = repo.Create(ctx, pt)

		pt.Amount = 75.0
		pt.Status = entity.PlannedTransactionStatusMatched
		err := repo.Update(ctx, pt)
		if err != nil {
			t.Fatalf("Update: unexpected error: %v", err)
		}

		found, _ := repo.GetByID(ctx, pt.ID, userID)
		if found.Amount != 75.0 || found.Status != entity.PlannedTransactionStatusMatched {
			t.Errorf("Update failed. Got amount %f, status %s", found.Amount, found.Status)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			ID:          uuid.New(),
			UserID:      userID,
			Amount:      10.0,
			Date:        time.Now().UTC(),
			Description: "To Delete",
		}
		_ = repo.Create(ctx, pt)

		err := repo.Delete(ctx, pt.ID, userID)
		if err != nil {
			t.Fatalf("Delete: unexpected error: %v", err)
		}

		_, err = repo.GetByID(ctx, pt.ID, userID)
		if err == nil {
			t.Error("GetByID: expected error after delete, got nil")
		}
	})

	t.Run("FindByUserID", func(t *testing.T) {
		_ = repo.Create(ctx, &entity.PlannedTransaction{UserID: userID, Amount: 1, Date: time.Now().UTC(), Description: "1"})
		_ = repo.Create(ctx, &entity.PlannedTransaction{UserID: userID, Amount: 2, Date: time.Now().UTC(), Description: "2"})
		_ = repo.Create(ctx, &entity.PlannedTransaction{UserID: otherUserID, Amount: 3, Date: time.Now().UTC(), Description: "3"})

		pts, err := repo.FindByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("FindByUserID: unexpected error: %v", err)
		}

		if len(pts) < 2 {
			t.Errorf("FindByUserID: expected at least 2 PTs, got %d", len(pts))
		}
		for _, p := range pts {
			if p.UserID != userID {
				t.Errorf("FindByUserID: found PT belonging to other user: %v", p.UserID)
			}
		}
	})
}
