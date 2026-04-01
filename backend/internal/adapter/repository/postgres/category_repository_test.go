package postgres

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestCategoryRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	repo := NewCategoryRepository(globalPool, setupLogger())

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'cat_user', 'hash', 'cat@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	otherUserID := uuid.New()
	_, err = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'other_cat_user', 'hash', 'other_cat@example.com')", otherUserID)
	if err != nil {
		t.Fatalf("failed to insert other user: %v", err)
	}

	t.Run("Save_and_FindByID", func(t *testing.T) {
		catID := uuid.New()
		cat := entity.Category{
			ID:     catID,
			UserID: userID,
			Name:   "Test Category",
			Color:  "#123456",
		}

		saved, err := repo.Save(ctx, cat)
		if err != nil {
			t.Fatalf("Save: unexpected error: %v", err)
		}
		if saved.Name != "Test Category" {
			t.Errorf("name: want 'Test Category', got '%s'", saved.Name)
		}

		found, err := repo.FindByID(ctx, catID, userID)
		if err != nil {
			t.Fatalf("FindByID: unexpected error: %v", err)
		}
		if found.ID != catID {
			t.Errorf("id: want %v, got %v", catID, found.ID)
		}
	})

	t.Run("FindByID_Isolation", func(t *testing.T) {
		catID := uuid.New()
		_, _ = repo.Save(ctx, entity.Category{
			ID:     catID,
			UserID: userID,
			Name:   "User's Private Category",
		})

		// Try to find as other user
		_, err := repo.FindByID(ctx, catID, otherUserID)
		if err == nil {
			t.Error("FindByID: expected error for other user, got nil")
		}
	})

	t.Run("FindAll_Isolation", func(t *testing.T) {
		_, _ = repo.Save(ctx, entity.Category{UserID: userID, Name: "User Cat 1"})
		_, _ = repo.Save(ctx, entity.Category{UserID: userID, Name: "User Cat 2"})
		_, _ = repo.Save(ctx, entity.Category{UserID: otherUserID, Name: "Other User Cat"})

		cats, err := repo.FindAll(ctx, userID)
		if err != nil {
			t.Fatalf("FindAll: unexpected error: %v", err)
		}

		if len(cats) < 2 {
			t.Errorf("FindAll: expected at least 2 categories, got %d", len(cats))
		}

		for _, c := range cats {
			if c.UserID != userID {
				t.Errorf("FindAll: found category belonging to other user: %v", c.UserID)
			}
		}
	})

	t.Run("Save_Upsert_Conflict_On_Name_Per_User", func(t *testing.T) {
		name := "Duplicate Name"
		cat1, _ := repo.Save(ctx, entity.Category{UserID: userID, Name: name, Color: "#111111"})
		cat2, _ := repo.Save(ctx, entity.Category{UserID: userID, Name: name, Color: "#222222"})

		if cat1.ID != cat2.ID {
			t.Errorf("Upsert: expected same ID for duplicate name, got %v and %v", cat1.ID, cat2.ID)
		}

		// Other user should be able to use the same name
		catOther, err := repo.Save(ctx, entity.Category{UserID: otherUserID, Name: name})
		if err != nil {
			t.Fatalf("Save other user same name: %v", err)
		}
		if catOther.ID == cat1.ID {
			t.Error("Save other user same name: expected different ID")
		}
	})

	t.Run("Update_Isolation", func(t *testing.T) {
		cat, _ := repo.Save(ctx, entity.Category{UserID: userID, Name: "Update Me"})
		
		cat.Name = "Updated"
		_, err := repo.Update(ctx, entity.Category{ID: cat.ID, UserID: otherUserID, Name: "Hijack"})
		if err == nil {
			t.Error("Update: expected error when updating other user's category, got nil")
		}
	})

	t.Run("Delete_Isolation", func(t *testing.T) {
		cat, _ := repo.Save(ctx, entity.Category{UserID: userID, Name: "Delete Me"})

		err := repo.Delete(ctx, cat.ID, otherUserID)
		if err == nil {
			t.Error("Delete: expected error when deleting other user's category, got nil")
		}

		err = repo.Delete(ctx, cat.ID, userID)
		if err != nil {
			t.Fatalf("Delete: unexpected error: %v", err)
		}
	})
}
