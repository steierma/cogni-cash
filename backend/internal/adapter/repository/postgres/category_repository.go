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

func (r *CategoryRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	r.Logger.Info("Finding all categories (including shared)", "user_id", userID)
	rows, err := r.pool.Query(ctx, `
		SELECT 
			c.id, c.user_id, c.name, c.color, c.is_variable_spending, c.forecast_strategy, c.created_at, c.deleted_at,
			EXISTS(SELECT 1 FROM shared_categories sc WHERE sc.category_id = c.id) as is_shared,
			COALESCE((SELECT array_agg(shared_with_user_id) FROM shared_categories WHERE category_id = c.id), '{}') as shared_with,
			c.user_id as owner_id
		FROM categories c
		WHERE (c.user_id = $1 
		   OR c.id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $1))
		   AND c.deleted_at IS NULL
		ORDER BY c.name ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("category repo: find all: %w", err)
	}
	defer rows.Close()

	cats := make([]entity.Category, 0)
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Color, &c.IsVariableSpending, &c.ForecastStrategy, &c.CreatedAt, &c.DeletedAt, &c.IsShared, &c.SharedWith, &c.OwnerID); err != nil {
			return nil, fmt.Errorf("category repo: scan: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Category, error) {
	r.Logger.Info("Finding category by ID", "id", id, "user_id", userID)
	var c entity.Category
	err := r.pool.QueryRow(ctx, `
		SELECT 
			c.id, c.user_id, c.name, c.color, c.is_variable_spending, c.forecast_strategy, c.created_at, c.deleted_at,
			EXISTS(SELECT 1 FROM shared_categories sc WHERE sc.category_id = c.id) as is_shared,
			COALESCE((SELECT array_agg(shared_with_user_id) FROM shared_categories WHERE category_id = c.id), '{}') as shared_with,
			c.user_id as owner_id
		FROM categories c 
		WHERE c.id = $1 AND (c.user_id = $2 OR c.id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $2))`, id, userID).
		Scan(&c.ID, &c.UserID, &c.Name, &c.Color, &c.IsVariableSpending, &c.ForecastStrategy, &c.CreatedAt, &c.DeletedAt, &c.IsShared, &c.SharedWith, &c.OwnerID)
	if err != nil {
		return entity.Category{}, fmt.Errorf("category repo: find by id: %w", err)
	}
	return c, nil
}

func (r *CategoryRepository) Save(ctx context.Context, cat entity.Category) (entity.Category, error) {
	r.Logger.Info("Saving category", "name", cat.Name, "user_id", cat.UserID)
	if cat.ID == uuid.Nil {
		cat.ID = uuid.New()
	}
	if cat.Color == "" {
		cat.Color = "#6366f1"
	}

	if cat.ForecastStrategy == "" {
		cat.ForecastStrategy = "3y"
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (id, user_id, name, color, is_variable_spending, forecast_strategy)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (name, user_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, user_id, name, color, is_variable_spending, forecast_strategy, created_at, deleted_at`,
		cat.ID, cat.UserID, cat.Name, cat.Color, cat.IsVariableSpending, cat.ForecastStrategy).
		Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Color, &cat.IsVariableSpending, &cat.ForecastStrategy, &cat.CreatedAt, &cat.DeletedAt)
	if err != nil {
		return entity.Category{}, fmt.Errorf("category repo: save: %w", err)
	}
	return cat, nil
}

func (r *CategoryRepository) Update(ctx context.Context, cat entity.Category) (entity.Category, error) {
	r.Logger.Info("Updating category", "id", cat.ID, "name", cat.Name, "user_id", cat.UserID)
	if cat.ForecastStrategy == "" {
		cat.ForecastStrategy = "3y"
	}
	err := r.pool.QueryRow(ctx, `
		UPDATE categories SET name = $1, color = $2, is_variable_spending = $3, deleted_at = $4, forecast_strategy = $5
		WHERE id = $6 AND user_id = $7
		RETURNING id, user_id, name, color, is_variable_spending, forecast_strategy, created_at, deleted_at`,
		cat.Name, cat.Color, cat.IsVariableSpending, cat.DeletedAt, cat.ForecastStrategy, cat.ID, cat.UserID).
		Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Color, &cat.IsVariableSpending, &cat.ForecastStrategy, &cat.CreatedAt, &cat.DeletedAt)
	if err != nil {
		return entity.Category{}, fmt.Errorf("category repo: update: %w", err)
	}
	return cat, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.Logger.Info("Soft-deleting category", "id", id, "user_id", userID)
	tag, err := r.pool.Exec(ctx, `UPDATE categories SET deleted_at = NOW() WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("category repo: delete (soft): %w", err)
	}
	if tag.RowsAffected() == 0 {
		r.Logger.Warn("Category not found for delete", "id", id, "user_id", userID)
		return fmt.Errorf("category repo: not found: %s", id)
	}
	return nil
}
