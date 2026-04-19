package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
)

type PlannedTransactionRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

func NewPlannedTransactionRepository(pool *pgxpool.Pool, logger *slog.Logger) *PlannedTransactionRepository {
	return &PlannedTransactionRepository{pool: pool, Logger: logger}
}

func (r *PlannedTransactionRepository) Create(ctx context.Context, pt *entity.PlannedTransaction) error {
	r.Logger.Info("Creating planned transaction", "user_id", pt.UserID, "amount", pt.Amount)
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	if pt.Status == "" {
		pt.Status = entity.PlannedTransactionStatusPending
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO planned_transactions (id, user_id, amount, date, description, category_id, status, matched_transaction_id, interval_months, end_date, is_superseded)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at`,
		pt.ID, pt.UserID, pt.Amount, pt.Date, pt.Description, pt.CategoryID, pt.Status, pt.MatchedTransactionID, pt.IntervalMonths, pt.EndDate, pt.IsSuperseded).
		Scan(&pt.CreatedAt)
	if err != nil {
		return fmt.Errorf("planned tx repo: create: %w", err)
	}
	return nil
}

func (r *PlannedTransactionRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PlannedTransaction, error) {
	r.Logger.Info("Getting planned transaction by id", "id", id, "user_id", userID)
	var pt entity.PlannedTransaction
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, amount, date, description, category_id, status, matched_transaction_id, interval_months, end_date, is_superseded, created_at
		FROM planned_transactions
		WHERE id = $1 AND user_id = $2`, id, userID).
		Scan(&pt.ID, &pt.UserID, &pt.Amount, &pt.Date, &pt.Description, &pt.CategoryID, &pt.Status, &pt.MatchedTransactionID, &pt.IntervalMonths, &pt.EndDate, &pt.IsSuperseded, &pt.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("planned tx repo: get by id: %w", err)
	}
	return &pt, nil
}

func (r *PlannedTransactionRepository) Update(ctx context.Context, pt *entity.PlannedTransaction) error {
	r.Logger.Info("Updating planned transaction", "id", pt.ID, "user_id", pt.UserID)
	tag, err := r.pool.Exec(ctx, `
		UPDATE planned_transactions
		SET amount = $1, date = $2, description = $3, category_id = $4, status = $5, matched_transaction_id = $6, interval_months = $7, end_date = $8, is_superseded = $9
		WHERE id = $10 AND user_id = $11`,
		pt.Amount, pt.Date, pt.Description, pt.CategoryID, pt.Status, pt.MatchedTransactionID, pt.IntervalMonths, pt.EndDate, pt.IsSuperseded, pt.ID, pt.UserID)
	if err != nil {
		return fmt.Errorf("planned tx repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("planned tx repo: update: not found or unauthorized")
	}
	return nil
}

func (r *PlannedTransactionRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.Logger.Info("Deleting planned transaction", "id", id, "user_id", userID)
	tag, err := r.pool.Exec(ctx, `DELETE FROM planned_transactions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("planned tx repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("planned tx repo: delete: not found or unauthorized")
	}
	return nil
}

func (r *PlannedTransactionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	r.Logger.Info("Finding planned transactions by user_id", "user_id", userID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, date, description, category_id, status, matched_transaction_id, interval_months, end_date, is_superseded, created_at
		FROM planned_transactions
		WHERE user_id = $1
		ORDER BY date ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("planned tx repo: find by user id: %w", err)
	}
	defer rows.Close()

	var pts []entity.PlannedTransaction
	for rows.Next() {
		var pt entity.PlannedTransaction
		if err := rows.Scan(&pt.ID, &pt.UserID, &pt.Amount, &pt.Date, &pt.Description, &pt.CategoryID, &pt.Status, &pt.MatchedTransactionID, &pt.IntervalMonths, &pt.EndDate, &pt.IsSuperseded, &pt.CreatedAt); err != nil {
			return nil, fmt.Errorf("planned tx repo: scan: %w", err)
		}
		pts = append(pts, pt)
	}
	return pts, rows.Err()
}

func (r *PlannedTransactionRepository) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	r.Logger.Info("Finding pending planned transactions by user_id", "user_id", userID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, date, description, category_id, status, matched_transaction_id, interval_months, end_date, is_superseded, created_at
		FROM planned_transactions
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY date ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("planned tx repo: find pending by user id: %w", err)
	}
	defer rows.Close()

	var pts []entity.PlannedTransaction
	for rows.Next() {
		var pt entity.PlannedTransaction
		if err := rows.Scan(&pt.ID, &pt.UserID, &pt.Amount, &pt.Date, &pt.Description, &pt.CategoryID, &pt.Status, &pt.MatchedTransactionID, &pt.IntervalMonths, &pt.EndDate, &pt.IsSuperseded, &pt.CreatedAt); err != nil {
			return nil, fmt.Errorf("planned tx repo: scan: %w", err)
		}
		pts = append(pts, pt)
	}
	return pts, rows.Err()
}
