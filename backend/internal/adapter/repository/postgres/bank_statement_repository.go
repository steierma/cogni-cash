package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
)

// BankStatementRepository implements port.BankStatementRepository using pgx.
type BankStatementRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

func NewBankStatementRepository(pool *pgxpool.Pool, logger *slog.Logger) *BankStatementRepository {
	return &BankStatementRepository{pool: pool, Logger: logger}
}

func (r *BankStatementRepository) Save(ctx context.Context, stmt entity.BankStatement) error {
	r.Logger.Info("Saving bank statement", "content_hash", stmt.ContentHash, "user_id", stmt.UserID)
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("bank_statement repo: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	stmtID := stmt.ID
	if stmtID == uuid.Nil {
		stmtID = uuid.New()
	}

	var statementDate *time.Time
	if !stmt.StatementDate.IsZero() {
		statementDate = &stmt.StatementDate
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO bank_statements
			(id, user_id, account_holder, iban,
			 statement_date, statement_no,
			 old_balance, new_balance, currency, original_file, content_hash, statement_type, bank_account_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12, $13)`,
		stmtID, stmt.UserID, stmt.AccountHolder, stmt.IBAN,
		statementDate, stmt.StatementNo, stmt.OldBalance, stmt.NewBalance,
		stmt.Currency, stmt.OriginalFile, stmt.ContentHash,
		string(stmt.StatementType), stmt.BankAccountID,
	)
	if err != nil {
		if isDuplicateHashError(err) {
			r.Logger.Warn("Duplicate bank statement detected", "content_hash", stmt.ContentHash, "user_id", stmt.UserID)
			return entity.ErrDuplicate
		}
		return fmt.Errorf("bank_statement repo: insert statement: %w", err)
	}

	batch := &pgx.Batch{}
	for _, t := range stmt.Transactions {
		var loc *string
		if t.Location != "" {
			loc = &t.Location
		}

		batch.Queue(`
			INSERT INTO transactions
			        (id, user_id, bank_statement_id, bank_account_id, booking_date, valuta_date,
			         description, location, amount, currency, base_amount, base_currency, transaction_type, reference, category_id, content_hash, statement_type, reviewed,
			         counterparty_name, counterparty_iban, bank_transaction_code, mandate_reference, is_payslip_verified, subscription_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20, $21, $22, $23, $24)
			ON CONFLICT (content_hash, user_id)
			DO UPDATE SET
				counterparty_name = EXCLUDED.counterparty_name,
				counterparty_iban = EXCLUDED.counterparty_iban,
				bank_transaction_code = EXCLUDED.bank_transaction_code,
				mandate_reference = EXCLUDED.mandate_reference,
				bank_account_id = COALESCE(transactions.bank_account_id, EXCLUDED.bank_account_id),
				location = COALESCE(transactions.location, EXCLUDED.location)`,
			uuid.New(),
			stmt.UserID,
			stmtID,
			stmt.BankAccountID,
			t.BookingDate,
			t.ValutaDate,
			t.Description,
			loc,
			t.Amount,
			t.Currency,
			t.BaseAmount,
			t.BaseCurrency,
			string(t.Type),
			t.Reference,
			t.CategoryID,
			t.ContentHash,
			string(t.StatementType),
			t.Reviewed,
			t.CounterpartyName,
			t.CounterpartyIban,
			t.BankTransactionCode,
			t.MandateReference,
			t.IsPayslipVerified,
			t.SubscriptionID,
		)
	}

	br := tx.SendBatch(ctx, batch)
	for range stmt.Transactions {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("bank_statement repo: insert transaction: %w", err)
		}
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("bank_statement repo: close batch: %w", err)
	}

	r.Logger.Info("Bank statement and transactions saved successfully", "statement_id", stmtID)
	return tx.Commit(ctx)
}

func (r *BankStatementRepository) CreateTransactions(ctx context.Context, txns []entity.Transaction) error {
	batch := &pgx.Batch{}
	for _, t := range txns {
		if t.ID == uuid.Nil {
			t.ID = uuid.New()
		}

		var loc *string
		if t.Location != "" {
			loc = &t.Location
		}

		batch.Queue(`
			INSERT INTO transactions
			        (id, user_id, bank_account_id, booking_date, valuta_date,
			         description, location, amount, currency, base_amount, base_currency, transaction_type, reference, category_id, content_hash, statement_type, reviewed,
			         counterparty_name, counterparty_iban, bank_transaction_code, mandate_reference, is_payslip_verified, subscription_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20, $21, $22, $23, $24)
			ON CONFLICT (content_hash, user_id) 
			DO UPDATE SET
				counterparty_name = EXCLUDED.counterparty_name,
				counterparty_iban = EXCLUDED.counterparty_iban,
				bank_transaction_code = EXCLUDED.bank_transaction_code,
				mandate_reference = EXCLUDED.mandate_reference,
				location = COALESCE(transactions.location, EXCLUDED.location)`,
			t.ID,
			t.UserID,
			t.BankAccountID,
			t.BookingDate,
			t.ValutaDate,
			t.Description,
			loc,
			t.Amount,
			t.Currency,
			t.BaseAmount,
			t.BaseCurrency,
			string(t.Type),
			t.Reference,
			t.CategoryID,
			t.ContentHash,
			string(t.StatementType),
			t.Reviewed,
			t.CounterpartyName,
			t.CounterpartyIban,
			t.BankTransactionCode,
			t.MandateReference,
			t.IsPayslipVerified,
			t.SubscriptionID,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range txns {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("bank_statement repo: create transactions: %w", err)
		}
	}
	return nil
}

func (r *BankStatementRepository) FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	r.Logger.Info("FindTransactions started", "user_id", filter.UserID, "include_shared", filter.IncludeShared, "cat_id", filter.CategoryID)

	query := `
	        SELECT t.id, t.user_id, t.bank_account_id, t.booking_date, t.valuta_date, t.description, t.location, t.amount,
	               t.currency, t.base_amount, t.base_currency, t.transaction_type, t.reference, t.category_id, t.content_hash,
	               t.is_reconciled, t.reconciliation_id, COALESCE(ba.account_type, t.statement_type, b.statement_type, 'giro'),
	               t.reviewed, t.counterparty_name, t.counterparty_iban, t.bank_transaction_code, t.mandate_reference, t.is_payslip_verified,
	               (t.user_id != $1) as is_shared,
	               t.user_id as owner_id,
	               t.subscription_id
	        FROM transactions t		LEFT JOIN bank_statements b ON t.bank_statement_id = b.id
		LEFT JOIN bank_accounts ba ON t.bank_account_id = ba.id`

	if filter.IncludeShared {
		// Include my transactions OR transactions in categories shared WITH me OR bank accounts shared WITH me
		query += ` WHERE (t.user_id = $1 
		   OR t.category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $1)
		   OR t.bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $1))`
	} else {
		// Only my transactions
		query += ` WHERE t.user_id = $1`
	}

	// Important: args[0] is always filter.UserID
	args := []any{filter.UserID}

	addCondition := func(condition string, arg any) {
		args = append(args, arg)
		query += fmt.Sprintf(" AND %s $%d", condition, len(args))
	}

	if filter.StatementID != nil {
		// If filtering by statement, we usually only want our own statements.
		// But for shared transactions, we might not have the statement ID.
		// So we only apply this if it's our own transaction or if we really want to restrict to that statement.
		addCondition("t.bank_statement_id =", *filter.StatementID)
	}
	if filter.CategoryID != nil {
		addCondition("t.category_id =", *filter.CategoryID)
	}
	if filter.Type == "credit" {
		query += " AND t.amount > 0"
	} else if filter.Type == "debit" {
		query += " AND t.amount <= 0"
	}
	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		idx := len(args)
		query += fmt.Sprintf(" AND (t.description ILIKE $%d OR t.reference ILIKE $%d OR t.location ILIKE $%d OR t.counterparty_name ILIKE $%d OR t.counterparty_iban ILIKE $%d OR t.bank_transaction_code ILIKE $%d OR t.mandate_reference ILIKE $%d)", idx, idx, idx, idx, idx, idx, idx)
	}
	if filter.FromDate != nil {
		addCondition("t.booking_date >=", *filter.FromDate)
	}
	if filter.ToDate != nil {
		addCondition("t.booking_date <=", *filter.ToDate)
	}
	if filter.MinAmount != nil {
		addCondition("t.amount >=", *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		addCondition("t.amount <=", *filter.MaxAmount)
	}

	if filter.IsReconciled != nil {
		if !*filter.IsReconciled {
			query += ` AND t.is_reconciled = false`
		} else {
			query += ` AND t.is_reconciled = true`
		}
	}
	if filter.Reviewed != nil {
		if !*filter.Reviewed {
			query += ` AND t.reviewed = false`
		} else {
			query += ` AND t.reviewed = true`
		}
	}
	if filter.StatementType != nil {
		addCondition("COALESCE(ba.account_type, t.statement_type, b.statement_type, 'giro') =", string(*filter.StatementType))
	}
	if filter.SubscriptionID != nil {
		addCondition("t.subscription_id =", *filter.SubscriptionID)
	}

	query += " ORDER BY t.booking_date DESC, t.id"

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if filter.Offset > 0 {
		args = append(args, filter.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	r.Logger.Info("Executing FindTransactions query", "query", query, "args", args)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("bank_statement repo: find transactions: %w", err)
	}
	defer rows.Close()

	txns := make([]entity.Transaction, 0)
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("bank_statement repo: scan filtered transaction: %w", err)
		}
		txns = append(txns, t)
	}

	r.Logger.Info("FindTransactions finished", "count", len(txns))

	return txns, rows.Err()
}

func (r *BankStatementRepository) loadTransactions(ctx context.Context, stmtID uuid.UUID, userID uuid.UUID) ([]entity.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
	        SELECT t.id, t.user_id, t.bank_account_id, t.booking_date, t.valuta_date, t.description, t.location, t.amount,
	               t.currency, t.base_amount, t.base_currency, t.transaction_type, t.reference, t.category_id, t.content_hash,
	               t.is_reconciled, t.reconciliation_id, COALESCE(ba.account_type, t.statement_type, b.statement_type, 'giro'),
	               t.reviewed, t.counterparty_name, t.counterparty_iban, t.bank_transaction_code, t.mandate_reference, t.is_payslip_verified,
	               (t.user_id != $2 OR EXISTS(SELECT 1 FROM shared_categories WHERE category_id = t.category_id AND shared_with_user_id = $2) OR EXISTS(SELECT 1 FROM shared_bank_accounts WHERE bank_account_id = t.bank_account_id AND shared_with_user_id = $2)) as is_shared,
	               t.user_id as owner_id,
	               t.subscription_id
	        FROM transactions t		JOIN bank_statements b ON t.bank_statement_id = b.id
		LEFT JOIN bank_accounts ba ON t.bank_account_id = ba.id
		WHERE t.bank_statement_id = $1 AND (t.user_id = $2 OR t.bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $2))
		ORDER BY t.booking_date DESC, t.id`, stmtID, userID)
	if err != nil {
		return nil, fmt.Errorf("bank_statement repo: load transactions: %w", err)
	}
	defer rows.Close()

	var txns []entity.Transaction
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("bank_statement repo: scan transaction: %w", err)
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

// scanTransaction scans a single transaction row. Shared by FindTransactions
// and loadTransactions to avoid duplicating the nullable-column logic.
func scanTransaction(row scanner) (entity.Transaction, error) {
	var t entity.Transaction
	var desc, loc, currency, baseCurrency, txType, ref, stmtType *string
	var cpName, cpIban, bankTxCode, mandateRef *string

	if err := row.Scan(
		&t.ID,
		&t.UserID,
		&t.BankAccountID,
		&t.BookingDate,
		&t.ValutaDate,
		&desc,
		&loc,
		&t.Amount,
		&currency,
		&t.BaseAmount,
		&baseCurrency,
		&txType,
		&ref,
		&t.CategoryID,
		&t.ContentHash,
		&t.IsReconciled,
		&t.ReconciliationID,
		&stmtType,
		&t.Reviewed,
		&cpName,
		&cpIban,
		&bankTxCode,
		&mandateRef,
		&t.IsPayslipVerified,
		&t.IsShared,
		&t.OwnerID,
		&t.SubscriptionID,
	); err != nil {
		return entity.Transaction{}, err
	}

	if desc != nil {
		t.Description = *desc
	}
	if loc != nil {
		t.Location = *loc
	}
	if currency != nil {
		t.Currency = *currency
	}
	if baseCurrency != nil {
		t.BaseCurrency = *baseCurrency
	}
	if ref != nil {
		t.Reference = *ref
	}
	if txType != nil {
		t.Type = entity.TransactionType(*txType)
	}
	if stmtType != nil {
		t.StatementType = entity.StatementType(*stmtType)
	}
	if cpName != nil {
		t.CounterpartyName = *cpName
	}
	if cpIban != nil {
		t.CounterpartyIban = *cpIban
	}
	if bankTxCode != nil {
		t.BankTransactionCode = *bankTxCode
	}
	if mandateRef != nil {
		t.MandateReference = *mandateRef
	}
	return t, nil
}

func isDuplicateHashError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "content_hash")
	}
	return false
}

func (r *BankStatementRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BankStatement, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, account_holder, iban,
		       statement_date, statement_no,
		       old_balance, new_balance, currency, original_file, imported_at, statement_type, bank_account_id
		FROM bank_statements 
		WHERE id = $1 AND (user_id = $2 OR bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $2))`, id, userID)

	stmt, err := scanStatement(row)
	if err != nil {
		return entity.BankStatement{}, fmt.Errorf("bank_statement repo: find by id: %w", err)
	}

	stmt.Transactions, err = r.loadTransactions(ctx, id, userID)
	if err != nil {
		return entity.BankStatement{}, err
	}
	return stmt, nil
}

func (r *BankStatementRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BankStatement, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, account_holder, iban,
		       statement_date, statement_no,
		       old_balance, new_balance, currency, NULL::bytea as original_file, imported_at, statement_type, bank_account_id
		FROM bank_statements
		WHERE user_id = $1 OR bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $1)
		ORDER BY statement_date DESC NULLS LAST, imported_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("bank_statement repo: find all: %w", err)
	}
	defer rows.Close()

	var stmts []entity.BankStatement
	for rows.Next() {
		s, err := scanStatement(rows)
		if err != nil {
			return nil, fmt.Errorf("bank_statement repo: scan row: %w", err)
		}
		stmts = append(stmts, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range stmts {
		stmts[i].Transactions, err = r.loadTransactions(ctx, stmts[i].ID, userID)
		if err != nil {
			return nil, fmt.Errorf("bank_statement repo: load transactions for %s: %w", stmts[i].ID, err)
		}
	}
	return stmts, nil
}

func (r *BankStatementRepository) FindSummaries(ctx context.Context, userID uuid.UUID) ([]entity.BankStatementSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			bs.id,
			bs.statement_no,
			bs.iban,
			bs.currency,
			bs.new_balance,
			bs.statement_date,
			bs.statement_type,
			MIN(t.booking_date) as start_date,
			MAX(t.booking_date) as end_date,
			COUNT(t.id) as transaction_count,
			(bs.original_file IS NOT NULL AND length(bs.original_file) > 0) AS has_original_file
		FROM bank_statements bs
		LEFT JOIN transactions t ON t.bank_statement_id = bs.id
		WHERE bs.user_id = $1 OR bs.bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $1)
		GROUP BY bs.id
		ORDER BY end_date DESC NULLS LAST, bs.statement_date DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("bank_statement repo: find summaries: %w", err)
	}
	defer rows.Close()

	var summaries []entity.BankStatementSummary
	for rows.Next() {
		var s entity.BankStatementSummary
		var stmtDate, startDate, endDate *time.Time
		// Use pointers to safely handle possible database NULLs left by legacy/dummy tests
		var iban, currency, stmtType *string

		if err := rows.Scan(
			&s.ID,
			&s.StatementNo,
			&iban,
			&currency,
			&s.NewBalance,
			&stmtDate,
			&stmtType,
			&startDate,
			&endDate,
			&s.TransactionCount,
			&s.HasOriginalFile,
		); err != nil {
			return nil, fmt.Errorf("bank_statement repo: scan summary: %w", err)
		}

		if iban != nil {
			s.IBAN = *iban
		}
		if currency != nil {
			s.Currency = *currency
		}
		if stmtType != nil {
			s.StatementType = entity.StatementType(*stmtType)
		}

		if stmtDate != nil {
			s.StartDate = *stmtDate
			s.EndDate = *stmtDate
		}
		if startDate != nil {
			s.StartDate = *startDate
		}
		if endDate != nil {
			s.EndDate = *endDate
		}

		s.PeriodLabel = s.StartDate.Format("Jan 2006")
		if s.StartDate.Month() != s.EndDate.Month() || s.StartDate.Year() != s.EndDate.Year() {
			s.PeriodLabel = s.StartDate.Format("Jan 2006") + " - " + s.EndDate.Format("Jan 2006")
		}

		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
}

func (r *BankStatementRepository) SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	return r.FindTransactions(ctx, filter)
}

func (r *BankStatementRepository) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, examplesCount int) ([]entity.CategorizationExample, error) {
	// Fetch a total of examplesCount unique transaction combinations across all categories,
	// prioritizing most recently imported ones.
	query := `
		WITH unique_txns AS (
			SELECT DISTINCT ON (description, counterparty_name, category_id)
				c.name as category_name,
				t.description,
				t.reference,
				t.counterparty_name,
				t.counterparty_iban,
				t.bank_transaction_code,
				t.mandate_reference,
				t.booking_date
			FROM transactions t
			JOIN categories c ON t.category_id = c.id
			WHERE t.user_id = $2 AND t.category_id IS NOT NULL
			ORDER BY description, counterparty_name, category_id, t.booking_date DESC
		)
		SELECT 
			category_name, 
			description, 
			reference, 
			counterparty_name, 
			counterparty_iban, 
			bank_transaction_code, 
			mandate_reference
		FROM unique_txns
		ORDER BY booking_date DESC
		LIMIT $1`

	rows, err := r.pool.Query(ctx, query, examplesCount, userID)
	if err != nil {
		return nil, fmt.Errorf("bank_statement repo: get examples: %w", err)
	}
	defer rows.Close()

	var examples []entity.CategorizationExample
	for rows.Next() {
		var ex entity.CategorizationExample
		var cpName, cpIban, bankTxCode, mandateRef *string
		if err := rows.Scan(
			&ex.Category,
			&ex.Description,
			&ex.Reference,
			&cpName,
			&cpIban,
			&bankTxCode,
			&mandateRef,
		); err != nil {
			return nil, err
		}
		if cpName != nil {
			ex.CounterpartyName = *cpName
		}
		if cpIban != nil {
			ex.CounterpartyIban = *cpIban
		}
		if bankTxCode != nil {
			ex.BankTransactionCode = *bankTxCode
		}
		if mandateRef != nil {
			ex.MandateReference = *mandateRef
		}
		examples = append(examples, ex)
	}
	return examples, rows.Err()
}

func (r *BankStatementRepository) FindMatchingCategory(ctx context.Context, userID uuid.UUID, txn port.TransactionToCategorize) (*uuid.UUID, error) {
	// 1. Try exact match on description + counterparty_name
	queryExact := `
		SELECT category_id 
		FROM transactions 
		WHERE user_id = $1 AND category_id IS NOT NULL 
		  AND description = $2 
		  AND (counterparty_name = $3 OR (counterparty_name IS NULL AND $3 = ''))
		LIMIT 1`

	var catID uuid.UUID
	err := r.pool.QueryRow(ctx, queryExact, userID, txn.Description, txn.CounterpartyName).Scan(&catID)
	if err == nil {
		return &catID, nil
	}

	// 2. Try fuzzy match with similarity threshold (e.g., 0.65 as requested)
	// We prioritize the same counterparty first.
	queryFuzzy := `
		SELECT category_id 
		FROM transactions 
		WHERE user_id = $1 AND category_id IS NOT NULL 
		  AND (
		  	SIMILARITY(description, $2) > 0.65 
		  	OR (counterparty_name IS NOT NULL AND $3 != '' AND SIMILARITY(counterparty_name, $3) > 0.65)
		  )
		ORDER BY SIMILARITY(description, $2) DESC, SIMILARITY(counterparty_name, $3) DESC
		LIMIT 1`

	err = r.pool.QueryRow(ctx, queryFuzzy, userID, txn.Description, txn.CounterpartyName).Scan(&catID)
	if err == nil {
		return &catID, nil
	}

	return nil, nil // No match found
}
func (r *BankStatementRepository) UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE transactions
		SET category_id = $1, reviewed = true
		WHERE content_hash = $2 AND user_id = $3`,
		categoryID, hash, userID)
	if err != nil {
		return fmt.Errorf("bank_statement repo: update transaction category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bank_statement repo: transaction not found: %s", hash)
	}
	return nil
}

func (r *BankStatementRepository) UpdateTransactionSubscription(ctx context.Context, hash string, subID *uuid.UUID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE transactions
		SET subscription_id = $1
		WHERE content_hash = $2 AND user_id = $3`,
		subID, hash, userID)
	if err != nil {
		return fmt.Errorf("bank_statement repo: update transaction subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bank_statement repo: transaction not found: %s", hash)
	}
	return nil
}

func (r *BankStatementRepository) MarkTransactionReviewed(ctx context.Context, contentHash string, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE transactions
		SET reviewed = true
		WHERE content_hash = $1 AND user_id = $2`,
		contentHash, userID)
	if err != nil {
		return fmt.Errorf("bank_statement repo: mark reviewed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bank_statement repo: transaction not found for review: %s", contentHash)
	}
	return nil
}

func (r *BankStatementRepository) MarkTransactionsReviewedBulk(ctx context.Context, hashes []string, userID uuid.UUID) error {
	if len(hashes) == 0 {
		return nil
	}
	
	// Convert slice of strings to postgres array format or rely on ANY
	tag, err := r.pool.Exec(ctx, `
		UPDATE transactions
		SET reviewed = true
		WHERE content_hash = ANY($1) AND user_id = $2`,
		hashes, userID)
	if err != nil {
		return fmt.Errorf("bank_statement repo: bulk mark reviewed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Just log or ignore if none were found
		return nil
	}
	return nil
}

func (r *BankStatementRepository) MarkTransactionReconciled(ctx context.Context, contentHash string, reconciliationID uuid.UUID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE transactions
		SET is_reconciled = true, reconciliation_id = $1, reviewed = true
		WHERE content_hash = $2 AND user_id = $3`,
		reconciliationID, contentHash, userID)
	if err != nil {
		return fmt.Errorf("bank_statement repo: mark reconciled: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bank_statement repo: transaction not found for reconciliation: %s", contentHash)
	}
	return nil
}


func (r *BankStatementRepository) UpdateTransactionBaseAmount(ctx context.Context, hash string, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE transactions SET base_amount = $1, base_currency = $2
		WHERE content_hash = $3 AND user_id = $4`, baseAmount, baseCurrency, hash, userID)
	return err
}

func (r *BankStatementRepository) LinkTransactionToStatement(ctx context.Context, id uuid.UUID, statementID uuid.UUID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE transactions
		SET bank_statement_id = $1
		WHERE id = $2 AND user_id = $3`,
		statementID, id, userID)

	if err != nil {
		return fmt.Errorf("bank_statement repo: link transaction to statement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bank_statement repo: transaction not found for linking: %s", id)
	}
	return nil
}

func (r *BankStatementRepository) UpdateStatementAccount(ctx context.Context, statementID uuid.UUID, bankAccountID *uuid.UUID, userID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Update the statement itself
	_, err = tx.Exec(ctx, `
		UPDATE bank_statements
		SET bank_account_id = $1
		WHERE id = $2 AND user_id = $3`,
		bankAccountID, statementID, userID)
	if err != nil {
		return fmt.Errorf("bank_statement repo: update statement account: %w", err)
	}

	// 2. Cascade to all transactions belonging to this statement
	_, err = tx.Exec(ctx, `
		UPDATE transactions
		SET bank_account_id = $1
		WHERE bank_statement_id = $2 AND user_id = $3`,
		bankAccountID, statementID, userID)
	if err != nil {
		return fmt.Errorf("bank_statement repo: cascade statement account to transactions: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *BankStatementRepository) GetTransactionsByAccountID(ctx context.Context, bankAccountID uuid.UUID, userID uuid.UUID) ([]entity.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
	        SELECT t.id, t.user_id, t.bank_account_id, t.booking_date, t.valuta_date, t.description, t.location, t.amount,
	               t.currency, t.base_amount, t.base_currency, t.transaction_type, t.reference, t.category_id, t.content_hash,
	               t.is_reconciled, t.reconciliation_id, COALESCE(ba.account_type, t.statement_type, b.statement_type, 'giro'),
	               t.reviewed, t.counterparty_name, t.counterparty_iban, t.bank_transaction_code, t.mandate_reference, t.is_payslip_verified,
	               (t.user_id != $2) as is_shared,
	               t.user_id as owner_id,
	               t.subscription_id
	        FROM transactions t
		LEFT JOIN bank_statements b ON t.bank_statement_id = b.id
		LEFT JOIN bank_accounts ba ON t.bank_account_id = ba.id
		WHERE t.bank_account_id = $1 AND (t.user_id = $2 OR t.bank_account_id IN (SELECT bank_account_id FROM shared_bank_accounts WHERE shared_with_user_id = $2))
		ORDER BY t.booking_date DESC, t.id`, bankAccountID, userID)
	if err != nil {
		return nil, fmt.Errorf("bank_statement repo: get transactions by account: %w", err)
	}
	defer rows.Close()

	var txns []entity.Transaction
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

func (r *BankStatementRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM bank_statements WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("bank_statement repo: delete statement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bank_statement repo: delete: statement not found: %s", id)
	}

	r.Logger.Info("Deleted bank statement and cascaded transactions", "statement_id", id, "user_id", userID)
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanStatement(row scanner) (entity.BankStatement, error) {
	var (
		s             entity.BankStatement
		statementDate *time.Time
		originalFile  []byte
		// Use pointers to scan safely in case the DB has NULLs (e.g. from tests)
		accHolder, iban, currency, stmtType *string
	)

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&accHolder,
		&iban,
		&statementDate,
		&s.StatementNo,
		&s.OldBalance,
		&s.NewBalance,
		&currency,
		&originalFile,
		&s.ImportedAt,
		&stmtType,
		&s.BankAccountID,
	)
	if err != nil {
		return entity.BankStatement{}, err
	}

	if accHolder != nil {
		s.AccountHolder = *accHolder
	}
	if iban != nil {
		s.IBAN = *iban
	}
	if currency != nil {
		s.Currency = *currency
	}
	if stmtType != nil {
		s.StatementType = entity.StatementType(*stmtType)
	}
	if statementDate != nil {
		s.StatementDate = *statementDate
	}
	s.OriginalFile = originalFile

	return s, nil
}
