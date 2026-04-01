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
	r.Logger.Info("Saving bank statement", "content_hash", stmt.ContentHash, "source_file", stmt.SourceFile, "user_id", stmt.UserID)
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
			(id, user_id, account_holder, iban, bic, account_number,
			 statement_date, statement_no,
			 old_balance, new_balance, currency, source_file, original_file, content_hash, statement_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14, $15)`,
		stmtID, stmt.UserID, stmt.AccountHolder, stmt.IBAN, stmt.BIC, stmt.AccountNumber,
		statementDate, stmt.StatementNo, stmt.OldBalance, stmt.NewBalance,
		stmt.Currency, stmt.SourceFile, stmt.OriginalFile, stmt.ContentHash,
		string(stmt.StatementType),
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
			        (id, user_id, bank_statement_id, booking_date, valuta_date,
			         description, location, amount, currency, transaction_type, reference, category_id, content_hash, statement_type, reviewed,
			         counterparty_name, counterparty_iban, bank_transaction_code, mandate_reference,
			         exchange_rate, amount_base_currency)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
			ON CONFLICT (content_hash, user_id) DO NOTHING`,
			uuid.New(),
			stmt.UserID,
			stmtID,
			t.BookingDate,
			t.ValutaDate,
			t.Description,
			loc,
			t.Amount,
			t.Currency,
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
			t.ExchangeRate,
			t.AmountBaseCurrency,
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
			         description, location, amount, currency, transaction_type, reference, category_id, content_hash, statement_type, reviewed,
			         counterparty_name, counterparty_iban, bank_transaction_code, mandate_reference,
			         exchange_rate, amount_base_currency)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
			ON CONFLICT (content_hash, user_id) DO NOTHING`,
			t.ID,
			t.UserID,
			t.BankAccountID,
			t.BookingDate,
			t.ValutaDate,
			t.Description,
			loc,
			t.Amount,
			t.Currency,
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
			t.ExchangeRate,
			t.AmountBaseCurrency,
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
	query := `
	        SELECT t.id, t.user_id, t.booking_date, t.valuta_date, t.description, t.location, t.amount,
	               t.currency, t.transaction_type, t.reference, t.category_id, t.content_hash,
	               t.is_reconciled, t.reconciliation_id, COALESCE(ba.account_type, t.statement_type, b.statement_type, 'giro'),
	               t.reviewed, t.counterparty_name, t.counterparty_iban, t.bank_transaction_code, t.mandate_reference,
	               COALESCE(t.exchange_rate, 1.0), t.amount_base_currency
	        FROM transactions t		LEFT JOIN bank_statements b ON t.bank_statement_id = b.id
		LEFT JOIN bank_accounts ba ON t.bank_account_id = ba.id
		WHERE t.user_id = $1`

	args := []any{filter.UserID}

	addCondition := func(condition string, arg any) {
		args = append(args, arg)
		query += fmt.Sprintf(" AND %s $%d", condition, len(args))
	}

	if filter.StatementID != nil {
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

	query += " ORDER BY t.booking_date DESC, t.id"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("bank_statement repo: find transactions: %w", err)
	}
	defer rows.Close()

	var txns []entity.Transaction
	for rows.Next() {
		t, err := scanTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("bank_statement repo: scan filtered transaction: %w", err)
		}
		txns = append(txns, t)
	}

	return txns, rows.Err()
}

func (r *BankStatementRepository) loadTransactions(ctx context.Context, stmtID uuid.UUID, userID uuid.UUID) ([]entity.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
	        SELECT t.id, t.user_id, t.booking_date, t.valuta_date, t.description, t.location, t.amount,
	               t.currency, t.transaction_type, t.reference, t.category_id, t.content_hash,
	               t.is_reconciled, t.reconciliation_id, COALESCE(ba.account_type, t.statement_type, b.statement_type, 'giro'),
	               t.reviewed, t.counterparty_name, t.counterparty_iban, t.bank_transaction_code, t.mandate_reference,
	               COALESCE(t.exchange_rate, 1.0), t.amount_base_currency
	        FROM transactions t		JOIN bank_statements b ON t.bank_statement_id = b.id
		LEFT JOIN bank_accounts ba ON t.bank_account_id = ba.id
		WHERE t.bank_statement_id = $1 AND t.user_id = $2
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
        var desc, loc, currency, txType, ref, stmtType *string
        var cpName, cpIban, bankTxCode, mandateRef *string
        var amountBase *float64

        if err := row.Scan(
                &t.ID,
                &t.UserID,
                &t.BookingDate,
                &t.ValutaDate,
                &desc,
                &loc,
                &t.Amount,
                &currency,
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
                &t.ExchangeRate,
                &amountBase,
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
        if amountBase != nil {
                t.AmountBaseCurrency = *amountBase
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
		SELECT id, user_id, account_holder, iban, bic, account_number,
		       statement_date, statement_no,
		       old_balance, new_balance, currency, source_file, original_file, imported_at, statement_type
		FROM bank_statements WHERE id = $1 AND user_id = $2`, id, userID)

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
		SELECT id, user_id, account_holder, iban, bic, account_number,
		       statement_date, statement_no,
		       old_balance, new_balance, currency, source_file, NULL::bytea as original_file, imported_at, statement_type
		FROM bank_statements
		WHERE user_id = $1
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
		WHERE bs.user_id = $1
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

func (r *BankStatementRepository) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, examplesPerCategory int) ([]entity.CategorizationExample, error) {
	// For each category, get N unique transaction combinations.
	// We use a LATERAL join to perform this per-category.
	query := `
		SELECT 
			c.name, 
			t.description, 
			t.reference, 
			t.counterparty_name, 
			t.counterparty_iban, 
			t.bank_transaction_code, 
			t.mandate_reference
		FROM categories c
		CROSS JOIN LATERAL (
			SELECT DISTINCT 
				description, 
				reference, 
				counterparty_name, 
				counterparty_iban, 
				bank_transaction_code, 
				mandate_reference
			FROM transactions
			WHERE category_id = c.id AND user_id = $2
			LIMIT $1
		) t
		WHERE c.user_id = $2
		ORDER BY c.name, t.description`

	rows, err := r.pool.Query(ctx, query, examplesPerCategory, userID)
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

func (r *BankStatementRepository) UpdateTransactionCategory(ctx context.Context, contentHash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE transactions 
		SET category_id = $1 
		WHERE content_hash = $2 AND user_id = $3`,
		categoryID, contentHash, userID)

	if err != nil {
		return fmt.Errorf("bank_statement repo: update transaction category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bank_statement repo: transaction not found: %s", contentHash)
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
		accHolder, iban, bic, accNum, currency, srcFile, stmtType *string
	)

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&accHolder,
		&iban,
		&bic,
		&accNum,
		&statementDate,
		&s.StatementNo,
		&s.OldBalance,
		&s.NewBalance,
		&currency,
		&srcFile,
		&originalFile,
		&s.ImportedAt,
		&stmtType,
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
	if bic != nil {
		s.BIC = *bic
	}
	if accNum != nil {
		s.AccountNumber = *accNum
	}
	if currency != nil {
		s.Currency = *currency
	}
	if srcFile != nil {
		s.SourceFile = *srcFile
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
