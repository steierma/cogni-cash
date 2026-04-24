package memory

import (
	"context"
	"sync"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const maxCategories = 500

type CategoryRepository struct {
	mu         sync.RWMutex
	categories map[uuid.UUID]entity.Category
	order      []uuid.UUID
}

func NewCategoryRepository() *CategoryRepository {
	r := &CategoryRepository{
		categories: make(map[uuid.UUID]entity.Category),
		order:      make([]uuid.UUID, 0, maxCategories),
	}
	r.seedData()
	return r
}

func (r *CategoryRepository) seedData() {
	userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	cats := []struct {
		Name  string
		Color string
		IsVar bool
	}{
		{"Groceries", "#10b981", true},
		{"Rent", "#3b82f6", false},
		{"Utilities", "#f59e0b", false},
		{"Transport", "#6366f1", true},
		{"Entertainment", "#ec4899", true},
		{"Salary", "#22c55e", false},
		{"Bonus", "#84cc16", false},
		{"Insurance", "#8b5cf6", false},
		{"Taxes", "#ef4444", false},
		{"Savings", "#14b8a6", false},
		{"Dining Out", "#f43f5e", true},
		{"Subscriptions", "#a855f7", false},
	}

	for _, c := range cats {
		id := uuid.New()
		cat := entity.Category{
			ID:                 id,
			UserID:             userID,
			Name:               c.Name,
			Color:              c.Color,
			IsVariableSpending: c.IsVar,
			OwnerID:            userID,
			CreatedAt:          time.Now().Add(-365 * 24 * time.Hour),
		}
		r.categories[id] = cat
		r.order = append(r.order, id)
	}
}

func (r *CategoryRepository) Save(ctx context.Context, category entity.Category) (entity.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	// For memory repository, we should enforce name uniqueness per user for parity
	for id, c := range r.categories {
		if c.UserID == category.UserID && c.Name == category.Name && id != category.ID {
			// Update the existing one (ON CONFLICT DO UPDATE parity)
			r.categories[id] = category
			return category, nil
		}
	}

	if _, exists := r.categories[category.ID]; !exists {
		if len(r.order) >= maxCategories {
			// Evict oldest
			oldestID := r.order[0]
			delete(r.categories, oldestID)
			r.order = r.order[1:]
		}
		r.order = append(r.order, category.ID)
	}

	r.categories[category.ID] = category
	return category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category entity.Category) (entity.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.categories[category.ID]
	if !ok || old.UserID != category.UserID {
		return entity.Category{}, entity.ErrCategoryNotFound
	}
	r.categories[category.ID] = category
	return category, nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	category, ok := r.categories[id]
	if !ok || category.UserID != userID {
		return entity.Category{}, entity.ErrCategoryNotFound
	}
	return category, nil
}

func (r *CategoryRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var categories []entity.Category
	for _, c := range r.categories {
		if c.UserID == userID && c.DeletedAt == nil {
			categories = append(categories, c)
		}
	}
	return categories, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	category, ok := r.categories[id]
	if !ok || category.UserID != userID {
		return entity.ErrCategoryNotFound
	}
	now := time.Now()
	category.DeletedAt = &now
	r.categories[id] = category
	return nil
}

var _ port.CategoryRepository = (*CategoryRepository)(nil)
