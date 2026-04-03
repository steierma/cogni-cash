package postgres

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PayslipRepository struct {
	db *pgxpool.Pool
}

func NewPayslipRepository(db *pgxpool.Pool) *PayslipRepository {
	return &PayslipRepository{db: db}
}

// Save inserts a new parsed payslip along with its raw file content and bonuses.
func (r *PayslipRepository) Save(ctx context.Context, p *entity.Payslip) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("payslip repo begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert the main payslip
	insertPayslipSQL := `
		INSERT INTO payslips (
			user_id, original_file_name, original_file_content, content_hash,
			period_month_num, period_year, employer_name, tax_class, tax_id,
			gross_pay, net_pay, payout_amount, custom_deductions
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13
		) RETURNING id
	`

	var fileContent interface{}
	if len(p.OriginalFileContent) > 0 {
		fileContent = p.OriginalFileContent
	}

	var payslipID string
	err = tx.QueryRow(ctx, insertPayslipSQL,
		p.UserID, p.OriginalFileName, fileContent, p.ContentHash,
		p.PeriodMonthNum, p.PeriodYear, p.EmployerName, p.TaxClass, p.TaxID,
		p.GrossPay, p.NetPay, p.PayoutAmount, p.CustomDeductions,
	).Scan(&payslipID)

	if err != nil {
		return fmt.Errorf("payslip repo insert: %w", err)
	}
	p.ID = payslipID

	// 2. Insert Bonuses (if any)
	if len(p.Bonuses) > 0 {
		insertBonusSQL := `
			INSERT INTO payslip_bonuses (payslip_id, description, amount)
			VALUES ($1, $2, $3)
		`
		for _, b := range p.Bonuses {
			_, err = tx.Exec(ctx, insertBonusSQL, payslipID, b.Description, b.Amount)
			if err != nil {
				return fmt.Errorf("payslip repo insert bonus: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *PayslipRepository) ExistsByHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM payslips WHERE content_hash = $1 AND user_id = $2)`
	err := r.db.QueryRow(ctx, query, hash, userID).Scan(&exists)
	return exists, err
}

func (r *PayslipRepository) ExistsByOriginalFileName(ctx context.Context, originalFileName string, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM payslips WHERE original_file_name = $1 AND user_id = $2)`
	err := r.db.QueryRow(ctx, query, originalFileName, userID).Scan(&exists)
	return exists, err
}

func (r *PayslipRepository) FindAll(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error) {
	query := `
		SELECT id, user_id, period_month_num, period_year, employer_name, tax_class, tax_id, 
		       gross_pay, net_pay, payout_amount, custom_deductions, created_at,
		       original_file_name
		FROM payslips
		WHERE user_id = $1
	`
	args := []interface{}{filter.UserID}

	if filter.Employer != "" {
		query += " AND employer_name = $2"
		args = append(args, filter.Employer)
	}

	query += " ORDER BY period_year DESC, period_month_num DESC, created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payslips []entity.Payslip
	for rows.Next() {
		var p entity.Payslip
		err := rows.Scan(
			&p.ID, &p.UserID, &p.PeriodMonthNum, &p.PeriodYear, &p.EmployerName, &p.TaxClass, &p.TaxID,
			&p.GrossPay, &p.NetPay, &p.PayoutAmount, &p.CustomDeductions, &p.CreatedAt,
			&p.OriginalFileName,
		)
		if err != nil {
			return nil, err
		}
		p.Bonuses, err = r.findBonuses(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		payslips = append(payslips, p)
	}
	return payslips, nil
}

func (r *PayslipRepository) findBonuses(ctx context.Context, payslipID string) ([]entity.Bonus, error) {
	rows, err := r.db.Query(ctx,
		`SELECT description, amount FROM payslip_bonuses WHERE payslip_id = $1 ORDER BY created_at`,
		payslipID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entity.Bonus
	for rows.Next() {
		var b entity.Bonus
		if err := rows.Scan(&b.Description, &b.Amount); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, nil
}

func (r *PayslipRepository) FindByID(ctx context.Context, id string, userID uuid.UUID) (entity.Payslip, error) {
	query := `
		SELECT id, user_id, period_month_num, period_year, employer_name, tax_class, tax_id, 
		       gross_pay, net_pay, payout_amount, custom_deductions,
		       original_file_name
		FROM payslips WHERE id = $1 AND user_id = $2
	`
	var p entity.Payslip
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&p.ID, &p.UserID, &p.PeriodMonthNum, &p.PeriodYear, &p.EmployerName, &p.TaxClass, &p.TaxID,
		&p.GrossPay, &p.NetPay, &p.PayoutAmount, &p.CustomDeductions,
		&p.OriginalFileName,
	)
	if err != nil {
		return p, err
	}

	p.Bonuses, err = r.findBonuses(ctx, p.ID)
	return p, err
}

func (r *PayslipRepository) Update(ctx context.Context, p *entity.Payslip) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("payslip repo update begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE payslips
		SET period_month_num = $1, period_year = $2, employer_name = $3, tax_class = $4, tax_id = $5,
		    gross_pay = $6, net_pay = $7, payout_amount = $8, custom_deductions = $9
		WHERE id = $10 AND user_id = $11`,
		p.PeriodMonthNum, p.PeriodYear, p.EmployerName, p.TaxClass, p.TaxID,
		p.GrossPay, p.NetPay, p.PayoutAmount, p.CustomDeductions,
		p.ID, p.UserID,
	)
	if err != nil {
		return fmt.Errorf("payslip repo update exec: %w", err)
	}

	if len(p.OriginalFileContent) > 0 {
		_, err = tx.Exec(ctx, `
			UPDATE payslips
			SET original_file_name = $1, original_file_content = $2, content_hash = $3
			WHERE id = $4 AND user_id = $5`,
			p.OriginalFileName, p.OriginalFileContent, p.ContentHash,
			p.ID, p.UserID,
		)
		if err != nil {
			return fmt.Errorf("payslip repo update file exec: %w", err)
		}
	}

	if _, err = tx.Exec(ctx, `DELETE FROM payslip_bonuses WHERE payslip_id = $1`, p.ID); err != nil {
		return fmt.Errorf("payslip repo update delete bonuses: %w", err)
	}
	for _, b := range p.Bonuses {
		_, err = tx.Exec(ctx,
			`INSERT INTO payslip_bonuses (payslip_id, description, amount) VALUES ($1, $2, $3)`,
			p.ID, b.Description, b.Amount,
		)
		if err != nil {
			return fmt.Errorf("payslip repo update insert bonus: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PayslipRepository) Delete(ctx context.Context, id string, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM payslips WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *PayslipRepository) GetOriginalFile(ctx context.Context, id string, userID uuid.UUID) ([]byte, string, string, error) {
	var content []byte
	var filename string

	query := `SELECT original_file_content, original_file_name FROM payslips WHERE id = $1 AND user_id = $2`
	err := r.db.QueryRow(ctx, query, id, userID).Scan(&content, &filename)
	return content, "application/pdf", filename, err
}
