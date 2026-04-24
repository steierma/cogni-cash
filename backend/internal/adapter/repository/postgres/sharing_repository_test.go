package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharingRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	repo := NewSharingRepository(globalPool, setupLogger())

	ownerID := uuid.New()
	friendID := uuid.New()

	// Seed users
	_, _ = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'owner', 'hash', 'owner@example.com')", ownerID)
	_, _ = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'friend', 'hash', 'friend@example.com')", friendID)

	t.Run("BankAccountSharing", func(t *testing.T) {
		connID := uuid.New()
		_, _ = globalPool.Exec(ctx, "INSERT INTO bank_connections (id, user_id, institution_id, institution_name, requisition_id, reference_id) VALUES ($1, $2, 'inst', 'Bank', 'req_shr', 'ref_shr')", connID, ownerID)

		accID := uuid.New()
		_, err := globalPool.Exec(ctx, "INSERT INTO bank_accounts (id, connection_id, provider_account_id, user_id, name, iban, currency, balance) VALUES ($1, $2, $3, $4, 'Test Acc', 'DE123', 'EUR', 100)", accID, connID, "prov_shr", ownerID)
		require.NoError(t, err)

		err = repo.ShareBankAccount(ctx, accID, ownerID, friendID, "view")
		assert.NoError(t, err)

		shares, err := repo.ListBankAccountShares(ctx, accID, ownerID)
		assert.NoError(t, err)
		assert.Contains(t, shares, friendID)

		err = repo.RevokeBankAccountShare(ctx, accID, ownerID, friendID)
		assert.NoError(t, err)
	})

	t.Run("CategorySharing", func(t *testing.T) {
		catID := uuid.New()
		_, err := globalPool.Exec(ctx, "INSERT INTO categories (id, user_id, name, color) VALUES ($1, $2, 'Test Cat', '#ffffff')", catID, ownerID)
		require.NoError(t, err)

		err = repo.ShareCategory(ctx, catID, ownerID, friendID, "view")
		assert.NoError(t, err)

		shares, err := repo.ListShares(ctx, catID, ownerID)
		assert.NoError(t, err)
		assert.Contains(t, shares, friendID)

		err = repo.RevokeShare(ctx, catID, ownerID, friendID)
		assert.NoError(t, err)
	})

	t.Run("InvoiceSharing", func(t *testing.T) {
		invID := uuid.New()
		_, err := globalPool.Exec(ctx, "INSERT INTO invoices (id, user_id, vendor, amount, currency, invoice_date) VALUES ($1, $2, 'Vendor', 50, 'EUR', '2024-01-01')", invID, ownerID)
		require.NoError(t, err)

		err = repo.ShareInvoice(ctx, invID, ownerID, friendID, "view")
		assert.NoError(t, err)

		shares, err := repo.ListInvoiceShares(ctx, invID, ownerID)
		assert.NoError(t, err)
		assert.Contains(t, shares, friendID)

		err = repo.RevokeInvoiceShare(ctx, invID, ownerID, friendID)
		assert.NoError(t, err)
	})
}
