package postgres

import (
	"context"
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

func (r *BankRepository) GetConnection(ctx context.Context, id uuid.UUID) (*entity.BankConnection, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, institution_id, institution_name, requisition_id, reference_id, status, created_at, expires_at
		FROM bank_connections WHERE id = $1
	`, id)

	var conn entity.BankConnection
	var status string
	err := row.Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.InstitutionID, &conn.InstitutionName, &conn.RequisitionID, &conn.ReferenceID, &status, &conn.CreatedAt, &conn.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("bank repo: get connection: %w", err)
	}
	conn.Status = entity.ConnectionStatus(status)
	return &conn, nil
}

func (r *BankRepository) GetConnectionByRequisition(ctx context.Context, requisitionID string) (*entity.BankConnection, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, institution_id, institution_name, requisition_id, reference_id, status, created_at, expires_at
		FROM bank_connections WHERE requisition_id = $1
	`, requisitionID)

	var conn entity.BankConnection
	var status string
	err := row.Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.InstitutionID, &conn.InstitutionName, &conn.RequisitionID, &conn.ReferenceID, &status, &conn.CreatedAt, &conn.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
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

func (r *BankRepository) UpdateConnectionStatus(ctx context.Context, id uuid.UUID, status entity.ConnectionStatus) error {
	_, err := r.pool.Exec(ctx, "UPDATE bank_connections SET status = $1 WHERE id = $2", string(status), id)
	if err != nil {
		return fmt.Errorf("bank repo: update connection status: %w", err)
	}
	return nil
}

func (r *BankRepository) UpdateRequisitionID(ctx context.Context, id uuid.UUID, requisitionID string) error {
	_, err := r.pool.Exec(ctx, "UPDATE bank_connections SET requisition_id = $1 WHERE id = $2", requisitionID, id)
	if err != nil {
		return fmt.Errorf("bank repo: update requisition id: %w", err)
	}
	return nil
}

func (r *BankRepository) DeleteConnection(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM bank_connections WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("bank repo: delete connection: %w", err)
	}
	return nil
}

// Accounts

func (r *BankRepository) UpsertAccounts(ctx context.Context, accounts []entity.BankAccount) error {
	batch := &pgx.Batch{}
	for _, acc := range accounts {
		if acc.ID == uuid.Nil {
			acc.ID = uuid.New()
		}
		batch.Queue(`
			INSERT INTO bank_accounts (id, connection_id, provider_account_id, iban, name, currency, balance, last_synced_at, account_type)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (provider_account_id) DO UPDATE SET
				balance = EXCLUDED.balance,
				last_synced_at = EXCLUDED.last_synced_at,
				iban = CASE WHEN EXCLUDED.iban <> '' THEN EXCLUDED.iban ELSE bank_accounts.iban END,
				name = CASE WHEN EXCLUDED.name <> '' THEN EXCLUDED.name ELSE bank_accounts.name END,
				account_type = EXCLUDED.account_type
		`, acc.ID, acc.ConnectionID, acc.ProviderAccountID, acc.IBAN, acc.Name, acc.Currency, acc.Balance, acc.LastSyncedAt, string(acc.AccountType))
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

func (r *BankRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*entity.BankAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, connection_id, provider_account_id, iban, name, currency, balance, last_synced_at, account_type
		FROM bank_accounts WHERE id = $1
	`, id)

	var acc entity.BankAccount
	var accType string
	err := row.Scan(&acc.ID, &acc.ConnectionID, &acc.ProviderAccountID, &acc.IBAN, &acc.Name, &acc.Currency, &acc.Balance, &acc.LastSyncedAt, &accType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("bank repo: get account by id: %w", err)
	}
	acc.AccountType = entity.StatementType(accType)
	return &acc, nil
}

func (r *BankRepository) GetAccountsByConnectionID(ctx context.Context, connectionID uuid.UUID) ([]entity.BankAccount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, connection_id, provider_account_id, iban, name, currency, balance, last_synced_at, account_type
		FROM bank_accounts WHERE connection_id = $1
	`, connectionID)
	if err != nil {
		return nil, fmt.Errorf("bank repo: list accounts: %w", err)
	}
	defer rows.Close()

	var accs []entity.BankAccount
	for rows.Next() {
		var acc entity.BankAccount
		var accType string
		if err := rows.Scan(&acc.ID, &acc.ConnectionID, &acc.ProviderAccountID, &acc.IBAN, &acc.Name, &acc.Currency, &acc.Balance, &acc.LastSyncedAt, &accType); err != nil {
			return nil, fmt.Errorf("bank repo: scan account: %w", err)
		}
		acc.AccountType = entity.StatementType(accType)
		accs = append(accs, acc)
	}
	return accs, nil
}

func (r *BankRepository) GetAccountByProviderID(ctx context.Context, providerAccountID string) (*entity.BankAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, connection_id, provider_account_id, iban, name, currency, balance, last_synced_at, account_type
		FROM bank_accounts WHERE provider_account_id = $1
	`, providerAccountID)

	var acc entity.BankAccount
	var accType string
	err := row.Scan(&acc.ID, &acc.ConnectionID, &acc.ProviderAccountID, &acc.IBAN, &acc.Name, &acc.Currency, &acc.Balance, &acc.LastSyncedAt, &accType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("bank repo: get account by provider id: %w", err)
	}
	acc.AccountType = entity.StatementType(accType)
	return &acc, nil
}

func (r *BankRepository) UpdateAccountBalance(ctx context.Context, id uuid.UUID, balance float64, syncedAt interface{}) error {
	_, err := r.pool.Exec(ctx, "UPDATE bank_accounts SET balance = $1, last_synced_at = $2 WHERE id = $3", balance, syncedAt, id)
	if err != nil {
		return fmt.Errorf("bank repo: update account balance: %w", err)
	}
	return nil
}

func (r *BankRepository) UpdateAccountType(ctx context.Context, id uuid.UUID, accType entity.StatementType) error {
	_, err := r.pool.Exec(ctx, "UPDATE bank_accounts SET account_type = $1 WHERE id = $2", string(accType), id)
	if err != nil {
		return fmt.Errorf("bank repo: update account type: %w", err)
	}
	return nil
}
