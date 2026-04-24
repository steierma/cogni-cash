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
	if pt.SchedulingStrategy == "" {
		pt.SchedulingStrategy = entity.SchedulingStrategyFixedDay
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO planned_transactions (id, user_id, amount, currency, base_amount, base_currency, date, description, category_id, status, matched_transaction_id, interval_months, scheduling_strategy, end_date, bank_account_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING created_at`,
		pt.ID, pt.UserID, pt.Amount, pt.Currency, pt.BaseAmount, pt.BaseCurrency, pt.Date, pt.Description, pt.CategoryID, pt.Status, pt.MatchedTransactionID, pt.IntervalMonths, pt.SchedulingStrategy, pt.EndDate, pt.BankAccountID).
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
		SELECT 
			id, user_id, amount, currency, base_amount, base_currency, date, description, category_id, status, matched_transaction_id, interval_months, scheduling_strategy, end_date, created_at, bank_account_id,
			(user_id != $2) as is_shared,
			user_id as owner_id
		FROM planned_transactions
		WHERE id = $1 AND (user_id = $2 
		   OR bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $2)
		   OR category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $2))`, id, userID).
		Scan(&pt.ID, &pt.UserID, &pt.Amount, &pt.Currency, &pt.BaseAmount, &pt.BaseCurrency, &pt.Date, &pt.Description, &pt.CategoryID, &pt.Status, &pt.MatchedTransactionID, &pt.IntervalMonths, &pt.SchedulingStrategy, &pt.EndDate, &pt.CreatedAt, &pt.BankAccountID, &pt.IsShared, &pt.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("planned tx repo: get by id: %w", err)
	}
	return &pt, nil
}

func (r *PlannedTransactionRepository) Update(ctx context.Context, pt *entity.PlannedTransaction) error {
	r.Logger.Info("Updating planned transaction", "id", pt.ID, "user_id", pt.UserID)
	tag, err := r.pool.Exec(ctx, `
		UPDATE planned_transactions
		SET amount = $1, currency = $2, base_amount = $3, base_currency = $4, date = $5, description = $6, category_id = $7, status = $8, matched_transaction_id = $9, interval_months = $10, scheduling_strategy = $11, end_date = $12, bank_account_id = $13
		WHERE id = $14 AND user_id = $15`,
		pt.Amount, pt.Currency, pt.BaseAmount, pt.BaseCurrency, pt.Date, pt.Description, pt.CategoryID, pt.Status, pt.MatchedTransactionID, pt.IntervalMonths, pt.SchedulingStrategy, pt.EndDate, pt.BankAccountID, pt.ID, pt.UserID)
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
		SELECT 
			pt.id, pt.user_id, pt.amount, pt.currency, pt.base_amount, pt.base_currency, pt.date, pt.description, pt.category_id, pt.status, pt.matched_transaction_id, pt.interval_months, pt.scheduling_strategy, pt.end_date, pt.created_at, pt.bank_account_id,
			(pt.user_id != $1) as is_shared,
			pt.user_id as owner_id
		FROM planned_transactions pt
		WHERE pt.user_id = $1 
		   OR pt.bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $1)
		   OR pt.category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $1)
		ORDER BY pt.date ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("planned tx repo: find by user id: %w", err)
	}
	defer rows.Close()

	var pts []entity.PlannedTransaction
	for rows.Next() {
		var pt entity.PlannedTransaction
		if err := rows.Scan(&pt.ID, &pt.UserID, &pt.Amount, &pt.Currency, &pt.BaseAmount, &pt.BaseCurrency, &pt.Date, &pt.Description, &pt.CategoryID, &pt.Status, &pt.MatchedTransactionID, &pt.IntervalMonths, &pt.SchedulingStrategy, &pt.EndDate, &pt.CreatedAt, &pt.BankAccountID, &pt.IsShared, &pt.OwnerID); err != nil {
			return nil, fmt.Errorf("planned tx repo: scan: %w", err)
		}
		pts = append(pts, pt)
	}
	return pts, rows.Err()
}

func (r *PlannedTransactionRepository) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	r.Logger.Info("Finding pending planned transactions by user_id", "user_id", userID)
	rows, err := r.pool.Query(ctx, `
		SELECT 
			pt.id, pt.user_id, pt.amount, pt.currency, pt.base_amount, pt.base_currency, pt.date, pt.description, pt.category_id, pt.status, pt.matched_transaction_id, pt.interval_months, pt.scheduling_strategy, pt.end_date, pt.created_at, pt.bank_account_id,
			(pt.user_id != $1) as is_shared,
			pt.user_id as owner_id
		FROM planned_transactions pt
		WHERE (pt.user_id = $1 
		   OR pt.bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $1)
		   OR pt.category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $1)) 
		   AND pt.status = 'pending'
		ORDER BY pt.date ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("planned tx repo: find pending by user id: %w", err)
	}
	defer rows.Close()

	var pts []entity.PlannedTransaction
	for rows.Next() {
		var pt entity.PlannedTransaction
		if err := rows.Scan(&pt.ID, &pt.UserID, &pt.Amount, &pt.Currency, &pt.BaseAmount, &pt.BaseCurrency, &pt.Date, &pt.Description, &pt.CategoryID, &pt.Status, &pt.MatchedTransactionID, &pt.IntervalMonths, &pt.SchedulingStrategy, &pt.EndDate, &pt.CreatedAt, &pt.BankAccountID, &pt.IsShared, &pt.OwnerID); err != nil {
			return nil, fmt.Errorf("planned tx repo: scan: %w", err)
		}
		pts = append(pts, pt)
	}
	return pts, rows.Err()
}
