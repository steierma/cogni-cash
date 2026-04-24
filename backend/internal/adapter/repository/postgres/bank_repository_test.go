package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogni-cash/internal/domain/entity"
)

func TestBankRepository_Integration(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	repo := NewBankRepository(globalPool, setupLogger())

	ownerID := uuid.New()
	partnerID := uuid.New()
	_, _ = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'owner', 'hash', 'owner@bank.com')", ownerID)
	_, _ = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'partner', 'hash', 'partner@bank.com')", partnerID)

	// Need a connection for non-virtual accounts
	connID := uuid.New()
	_, _ = globalPool.Exec(ctx, "INSERT INTO bank_connections (id, user_id, institution_id, institution_name, requisition_id, reference_id) VALUES ($1, $2, 'inst', 'Bank', 'req', 'ref')", connID, ownerID)

	t.Run("GetAccountsByUserID - Includes Shared", func(t *testing.T) {
		// 1. Create owner's account
		accID := uuid.New()
		_, err := globalPool.Exec(ctx, "INSERT INTO bank_accounts (id, connection_id, provider_account_id, user_id, name, iban, currency, balance, last_synced_at) VALUES ($1, $2, $3, $4, 'Owner Acc', 'IBAN1', 'EUR', 100, NOW())", accID, connID, "prov1", ownerID)
		require.NoError(t, err)

		// 2. Share it with partner
		_, err = globalPool.Exec(ctx, "INSERT INTO shared_bank_accounts (bank_account_id, owner_user_id, shared_with_user_id, permission_level) VALUES ($1, $2, $3, 'view')", accID, ownerID, partnerID)
		require.NoError(t, err)

		// 3. Get partner's accounts -> should see the shared one
		accounts, err := repo.GetAccountsByUserID(ctx, partnerID)
		assert.NoError(t, err)
		assert.Len(t, accounts, 1)
		assert.Equal(t, accID, accounts[0].ID)
		assert.True(t, accounts[0].IsShared)
		assert.Equal(t, ownerID, accounts[0].OwnerID)
	})

	t.Run("UpdateAccountType", func(t *testing.T) {
		accID := uuid.New()
		_, _ = globalPool.Exec(ctx, "INSERT INTO bank_accounts (id, connection_id, provider_account_id, user_id, name, iban, currency, balance, last_synced_at) VALUES ($1, $2, $3, $4, 'Acc', 'IBAN2', 'EUR', 0, NOW())", accID, connID, "prov2", ownerID)

		err := repo.UpdateAccountType(ctx, accID, entity.StatementTypeCreditCard, ownerID)
		assert.NoError(t, err)

		// Verify
		var accType string
		err = globalPool.QueryRow(ctx, "SELECT account_type FROM bank_accounts WHERE id = $1", accID).Scan(&accType)
		assert.NoError(t, err)
		assert.Equal(t, "credit_card", accType)
	})

	t.Run("DeleteConnection", func(t *testing.T) {
		cID := uuid.New()
		_, _ = globalPool.Exec(ctx, "INSERT INTO bank_connections (id, user_id, institution_id, institution_name, requisition_id, reference_id) VALUES ($1, $2, 'inst2', 'Bank2', 'req2', 'ref2')", cID, ownerID)

		err := repo.DeleteConnection(ctx, cID, ownerID)
		assert.NoError(t, err)

		var count int
		_ = globalPool.QueryRow(ctx, "SELECT COUNT(*) FROM bank_connections WHERE id = $1", cID).Scan(&count)
		assert.Equal(t, 0, count)
	})
}
