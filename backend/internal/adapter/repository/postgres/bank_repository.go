package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
)

type BankRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

func NewBankRepository(pool *pgxpool.Pool, logger *slog.Logger) *BankRepository {
	return &BankRepository{pool: pool, Logger: logger}
}

// Connections

func (r *BankRepository) CreateConnection(ctx context.Context, conn *entity.BankConnection) error {
	if conn.ID == uuid.Nil {
		conn.ID = uuid.New()
	}
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO bank_connections (id, user_id, provider, institution_id, institution_name, requisition_id, reference_id, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, conn.ID, conn.UserID, conn.Provider, conn.InstitutionID, conn.InstitutionName, conn.RequisitionID, conn.ReferenceID, string(conn.Status), conn.CreatedAt, conn.ExpiresAt)

	if err != nil {
		return fmt.Errorf("bank repo: create connection: %w", err)
	}
	return nil
}

func (r *BankRepository) GetConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.BankConnection, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, institution_id, institution_name, requisition_id, reference_id, status, created_at, expires_at
		FROM bank_connections WHERE id = $1 AND user_id = $2
	`, id, userID)

	var conn entity.BankConnection
	var status string
	err := row.Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.InstitutionID, &conn.InstitutionName, &conn.RequisitionID, &conn.ReferenceID, &status, &conn.CreatedAt, &conn.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("bank repo: get connection: %w", err)
	}
	conn.Status = entity.ConnectionStatus(status)
	return &conn, nil
}

func (r *BankRepository) GetConnectionByRequisition(ctx context.Context, requisitionID string, userID uuid.UUID) (*entity.BankConnection, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, institution_id, institution_name, requisition_id, reference_id, status, created_at, expires_at
		FROM bank_connections WHERE requisition_id = $1 AND user_id = $2
	`, requisitionID, userID)

	var conn entity.BankConnection
	var status string
	err := row.Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.InstitutionID, &conn.InstitutionName, &conn.RequisitionID, &conn.ReferenceID, &status, &conn.CreatedAt, &conn.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("bank repo: get connection by req: %w", err)
	}
	conn.Status = entity.ConnectionStatus(status)
	return &conn, nil
}

func (r *BankRepository) GetConnectionsByUserID(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, provider, institution_id, institution_name, requisition_id, reference_id, status, created_at, expires_at
		FROM bank_connections WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("bank repo: list connections: %w", err)
	}
	defer rows.Close()

	var conns []entity.BankConnection
	for rows.Next() {
		var conn entity.BankConnection
		var status string
		if err := rows.Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.InstitutionID, &conn.InstitutionName, &conn.RequisitionID, &conn.ReferenceID, &status, &conn.CreatedAt, &conn.ExpiresAt); err != nil {
			return nil, fmt.Errorf("bank repo: scan connection: %w", err)
		}
		conn.Status = entity.ConnectionStatus(status)
		conns = append(conns, conn)
	}
	return conns, nil
}

func (r *BankRepository) UpdateConnectionStatus(ctx context.Context, id uuid.UUID, status entity.ConnectionStatus, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE bank_connections SET status = $1 WHERE id = $2 AND user_id = $3", string(status), id, userID)
	if err != nil {
		return fmt.Errorf("bank repo: update connection status: %w", err)
	}
	return nil
}

func (r *BankRepository) UpdateRequisitionID(ctx context.Context, id uuid.UUID, requisitionID string, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE bank_connections SET requisition_id = $1 WHERE id = $2 AND user_id = $3", requisitionID, id, userID)
	if err != nil {
		return fmt.Errorf("bank repo: update requisition id: %w", err)
	}
	return nil
}

func (r *BankRepository) DeleteConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM bank_connections WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("bank repo: delete connection: %w", err)
	}
	return nil
}

// Accounts

func (r *BankRepository) UpsertAccounts(ctx context.Context, accounts []entity.BankAccount, userID uuid.UUID) error {
	batch := &pgx.Batch{}
	for _, acc := range accounts {
		if acc.ID == uuid.Nil {
			acc.ID = uuid.New()
		}
		batch.Queue(`
			INSERT INTO bank_accounts (id, connection_id, provider_account_id, iban, name, currency, balance, last_synced_at, account_type)
			SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
			FROM bank_connections
			WHERE id = $2 AND user_id = $10
			ON CONFLICT (provider_account_id) DO UPDATE SET
				balance = EXCLUDED.balance,
				last_synced_at = EXCLUDED.last_synced_at,
				iban = CASE WHEN EXCLUDED.iban <> '' THEN EXCLUDED.iban ELSE bank_accounts.iban END,
				name = CASE WHEN EXCLUDED.name <> '' THEN EXCLUDED.name ELSE bank_accounts.name END,
				account_type = EXCLUDED.account_type
		`, acc.ID, acc.ConnectionID, acc.ProviderAccountID, acc.IBAN, acc.Name, acc.Currency, acc.Balance, acc.LastSyncedAt, string(acc.AccountType), userID)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range accounts {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("bank repo: upsert accounts batch: %w", err)
		}
	}
	return nil
}

func (r *BankRepository) GetAccountByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.BankAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT ba.id, ba.connection_id, ba.provider_account_id, ba.iban, ba.name, ba.currency, ba.balance, ba.last_synced_at, ba.account_type, ba.last_sync_error
		FROM bank_accounts ba
		JOIN bank_connections bc ON ba.connection_id = bc.id
		WHERE ba.id = $1 AND bc.user_id = $2
	`, id, userID)

	var acc entity.BankAccount
	var accType string
	err := row.Scan(&acc.ID, &acc.ConnectionID, &acc.ProviderAccountID, &acc.IBAN, &acc.Name, &acc.Currency, &acc.Balance, &acc.LastSyncedAt, &accType, &acc.LastSyncError)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("bank repo: get account by id: %w", err)
	}
	acc.AccountType = entity.StatementType(accType)
	return &acc, nil
}

func (r *BankRepository) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]entity.BankAccount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ba.id, ba.connection_id, ba.provider_account_id, ba.iban, ba.name, ba.currency, ba.balance, ba.last_synced_at, ba.account_type, ba.last_sync_error
		FROM bank_accounts ba
		JOIN bank_connections bc ON ba.connection_id = bc.id
		WHERE bc.user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("bank repo: list accounts by user: %w", err)
	}
	defer rows.Close()

	var accs []entity.BankAccount
	for rows.Next() {
		var acc entity.BankAccount
		var accType string
		if err := rows.Scan(&acc.ID, &acc.ConnectionID, &acc.ProviderAccountID, &acc.IBAN, &acc.Name, &acc.Currency, &acc.Balance, &acc.LastSyncedAt, &accType, &acc.LastSyncError); err != nil {
			return nil, fmt.Errorf("bank repo: scan account: %w", err)
		}
		acc.AccountType = entity.StatementType(accType)
		accs = append(accs, acc)
	}
	return accs, nil
}

func (r *BankRepository) GetAccountsByConnectionID(ctx context.Context, connectionID uuid.UUID, userID uuid.UUID) ([]entity.BankAccount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ba.id, ba.connection_id, ba.provider_account_id, ba.iban, ba.name, ba.currency, ba.balance, ba.last_synced_at, ba.account_type, ba.last_sync_error
		FROM bank_accounts ba
		JOIN bank_connections bc ON ba.connection_id = bc.id
		WHERE ba.connection_id = $1 AND bc.user_id = $2
	`, connectionID, userID)
	if err != nil {
		return nil, fmt.Errorf("bank repo: list accounts by connection: %w", err)
	}
	defer rows.Close()

	var accs []entity.BankAccount
	for rows.Next() {
		var acc entity.BankAccount
		var accType string
		if err := rows.Scan(&acc.ID, &acc.ConnectionID, &acc.ProviderAccountID, &acc.IBAN, &acc.Name, &acc.Currency, &acc.Balance, &acc.LastSyncedAt, &accType, &acc.LastSyncError); err != nil {
			return nil, fmt.Errorf("bank repo: scan account: %w", err)
		}
		acc.AccountType = entity.StatementType(accType)
		accs = append(accs, acc)
	}
	return accs, nil
}

func (r *BankRepository) GetAccountByProviderID(ctx context.Context, providerAccountID string, userID uuid.UUID) (*entity.BankAccount, error) {
	query := `
		SELECT a.id, a.connection_id, a.provider_account_id, a.iban, a.name, 
		       a.currency, a.balance, a.last_synced_at, a.account_type, a.last_sync_error
		FROM bank_accounts a
		JOIN bank_connections c ON a.connection_id = c.id
		WHERE a.provider_account_id = $1 AND c.user_id = $2
	`
	var acc entity.BankAccount
	var accType string
	err := r.pool.QueryRow(ctx, query, providerAccountID, userID).Scan(
		&acc.ID, &acc.ConnectionID, &acc.ProviderAccountID, &acc.IBAN, &acc.Name,
		&acc.Currency, &acc.Balance, &acc.LastSyncedAt, &accType, &acc.LastSyncError,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil to match existing handling in this file
		}
		return nil, fmt.Errorf("bank repo: get account by provider id: %w", err)
	}
	acc.AccountType = entity.StatementType(accType)
	return &acc, nil
}

func (r *BankRepository) UpdateAccountBalance(ctx context.Context, id uuid.UUID, balance float64, syncedAt interface{}, errorMsg *string, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bank_accounts
		SET balance = $1, last_synced_at = $2, last_sync_error = $3
		FROM bank_connections bc
		WHERE bank_accounts.connection_id = bc.id
		  AND bank_accounts.id = $4
		  AND bc.user_id = $5
	`, balance, syncedAt, errorMsg, id, userID)
	if err != nil {
		return fmt.Errorf("bank repo: update account balance: %w", err)
	}
	return nil
}

func (r *BankRepository) UpdateAccountType(ctx context.Context, id uuid.UUID, accType entity.StatementType, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE bank_accounts
		SET account_type = $1
		FROM bank_connections bc
		WHERE bank_accounts.connection_id = bc.id
		  AND bank_accounts.id = $2
		  AND bc.user_id = $3
	`, string(accType), id, userID)
	if err != nil {
		return fmt.Errorf("bank repo: update account type: %w", err)
	}
	return nil
}
