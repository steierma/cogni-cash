package memory

import (
	"context"
	"sync"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type CategoryRepository struct {
	mu         sync.RWMutex
	categories map[uuid.UUID]entity.Category
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		categories: make(map[uuid.UUID]entity.Category),
	}
}

func (r *CategoryRepository) Save(ctx context.Context, category entity.Category) (entity.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}
	r.categories[category.ID] = category
	return category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category entity.Category) (entity.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.categories[category.ID]; !ok {
		return entity.Category{}, entity.ErrCategoryNotFound
	}
	r.categories[category.ID] = category
	return category, nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (entity.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	category, ok := r.categories[id]
	if !ok {
		return entity.Category{}, entity.ErrCategoryNotFound
	}
	return category, nil
}

func (r *CategoryRepository) FindAll(ctx context.Context) ([]entity.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var categories []entity.Category
	for _, c := range r.categories {
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.categories[id]; !ok {
		return entity.ErrCategoryNotFound
	}
	delete(r.categories, id)
	return nil
}

var _ port.CategoryRepository = (*CategoryRepository)(nil)
