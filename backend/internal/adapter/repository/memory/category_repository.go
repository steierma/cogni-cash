package memory

import (
	"context"
	"sync"

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
	return &CategoryRepository{
		categories: make(map[uuid.UUID]entity.Category),
		order:      make([]uuid.UUID, 0, maxCategories),
	}
}

func (r *CategoryRepository) Save(ctx context.Context, category entity.Category) (entity.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
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
		if c.UserID == userID {
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
	delete(r.categories, id)
	return nil
}

var _ port.CategoryRepository = (*CategoryRepository)(nil)
