package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
)

// CategoryRepository implements port.CategoryRepository using pgx.
type CategoryRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(pool *pgxpool.Pool, logger *slog.Logger) *CategoryRepository {
	return &CategoryRepository{pool: pool, Logger: logger}
}

func (r *CategoryRepository) FindAll(ctx context.Context) ([]entity.Category, error) {
	r.Logger.Info("Finding all categories")
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, color, created_at
		FROM categories
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("category repo: find all: %w", err)
	}
	defer rows.Close()

	var cats []entity.Category
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("category repo: scan: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (entity.Category, error) {
	r.Logger.Info("Finding category by ID", "id", id)
	var c entity.Category
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, color, created_at FROM categories WHERE id = $1`, id).
		Scan(&c.ID, &c.Name, &c.Color, &c.CreatedAt)
	if err != nil {
		return entity.Category{}, fmt.Errorf("category repo: find by id: %w", err)
	}
	return c, nil
}

func (r *CategoryRepository) Save(ctx context.Context, cat entity.Category) (entity.Category, error) {
	r.Logger.Info("Saving category", "name", cat.Name)
	if cat.ID == uuid.Nil {
		cat.ID = uuid.New()
	}
	if cat.Color == "" {
		cat.Color = "#6366f1"
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (id, name, color)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, color, created_at`,
		cat.ID, cat.Name, cat.Color).
		Scan(&cat.ID, &cat.Name, &cat.Color, &cat.CreatedAt)
	if err != nil {
		return entity.Category{}, fmt.Errorf("category repo: save: %w", err)
	}
	return cat, nil
}

func (r *CategoryRepository) Update(ctx context.Context, cat entity.Category) (entity.Category, error) {
	r.Logger.Info("Updating category", "id", cat.ID, "name", cat.Name)
	err := r.pool.QueryRow(ctx, `
		UPDATE categories SET name = $1, color = $2
		WHERE id = $3
		RETURNING id, name, color, created_at`,
		cat.Name, cat.Color, cat.ID).
		Scan(&cat.ID, &cat.Name, &cat.Color, &cat.CreatedAt)
	if err != nil {
		return entity.Category{}, fmt.Errorf("category repo: update: %w", err)
	}
	return cat, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.Logger.Info("Deleting category", "id", id)
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("category repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		r.Logger.Warn("Category not found for delete", "id", id)
		return fmt.Errorf("category repo: not found: %s", id)
	}
	return nil
}
