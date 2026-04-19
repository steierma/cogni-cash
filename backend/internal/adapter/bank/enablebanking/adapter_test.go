package enablebanking

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestAdapter(t *testing.T, handler http.HandlerFunc) (*Adapter, *httptest.Server) {
	server := httptest.NewServer(handler)
	
	// Generate a small RSA key for tests
	privKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	adapter := NewAdapter("test-app", privKey, logger)
	adapter.BaseURL = server.URL

	return adapter, server
}

func TestAdapter_FetchAccounts_CacheHit(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	sessionID := "test-session"

	adapter := &Adapter{
		accountsCache: make(map[string][]entity.BankAccount),
		logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	expectedAccounts := []entity.BankAccount{
		{ProviderAccountID: "acc-1", Name: "Giro"},
	}
	adapter.accountsCache[sessionID] = expectedAccounts

	accounts, err := adapter.FetchAccounts(ctx, userID, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)

	// Cache should be cleared after hit
	_, exists := adapter.accountsCache[sessionID]
	assert.False(t, exists)
}

func TestAdapter_FetchAccounts_Recovery(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	sessionID := "test-session"

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/sessions/"+sessionID, r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		resp := map[string]interface{}{
			"session_id": sessionID,
			"accounts": []map[string]interface{}{
				{
					"uid": "acc-123",
					"account_id": map[string]string{
						"iban": "DE12345",
					},
					"name":              "Test Giro",
					"currency":          "EUR",
					"cash_account_type": "CACC",
				},
				{
					"uid": "acc-456",
					"account_id": map[string]string{
						"iban": "DE67890",
					},
					"name":              "Test Credit",
					"currency":          "EUR",
					"cash_account_type": "CARD",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}

	adapter, server := setupTestAdapter(t, handler)
	defer server.Close()

	accounts, err := adapter.FetchAccounts(ctx, userID, sessionID)
	assert.NoError(t, err)
	assert.Len(t, accounts, 2)

	assert.Equal(t, "acc-123", accounts[0].ProviderAccountID)
	assert.Equal(t, entity.StatementTypeGiro, accounts[0].AccountType)
	assert.Equal(t, "acc-456", accounts[1].ProviderAccountID)
	assert.Equal(t, entity.StatementTypeCreditCard, accounts[1].AccountType)
}

func TestAdapter_FetchTransactions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	accID := "acc-123"
	dateFrom := time.Now().AddDate(0, 0, -7)

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/accounts/"+accID+"/transactions" {
			assert.Contains(t, r.URL.RawQuery, "date_from="+dateFrom.Format("2006-01-02"))
			
			resp := map[string]interface{}{
				"transactions": []map[string]interface{}{
					{
						"transaction_id": "tx-1",
						"booking_date":   "2026-04-10",
						"status":         "BOOK",
						"credit_debit_indicator": "DBIT",
						"transaction_amount": map[string]string{
							"amount":   "50.00",
							"currency": "EUR",
						},
						"remittance_information_unstructured": "Test Payment",
						"creditor_name": "Merchant A",
					},
					{
						"transaction_id": "tx-pending",
						"status":         "PDNG", // Should be skipped
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/accounts/"+accID+"/balances" {
			resp := map[string]interface{}{
				"balances": []map[string]interface{}{
					{
						"balance_amount": map[string]string{
							"amount":   "1234.56",
							"currency": "EUR",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	adapter, server := setupTestAdapter(t, handler)
	defer server.Close()

	txns, balance, err := adapter.FetchTransactions(ctx, userID, accID, &dateFrom, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1234.56, balance)
	assert.Len(t, txns, 1)
	assert.Equal(t, -50.00, txns[0].Amount)
	assert.Equal(t, entity.TransactionTypeDebit, txns[0].Type)
	assert.Equal(t, "Merchant A", txns[0].CounterpartyName)
}

func TestAdapter_FetchTransactions_RealING(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	accID := "acc-real-ing"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/accounts/"+accID+"/transactions" {
			resp := map[string]interface{}{
				"transactions": []map[string]interface{}{
					{
						"entry_reference": "000010000000001",
						"transaction_amount": map[string]string{
							"currency": "EUR",
							"amount":   "50.0",
						},
						"creditor": map[string]string{
							"name": "Generic Insurance Corp",
						},
						"creditor_account": map[string]string{
							"iban": "DE99123456789012345678",
						},
						"debtor_account": map[string]string{
							"iban": "DE88000000000000000000", // Anonymized User IBAN
						},
						"bank_transaction_code": map[string]string{
							"description": "Lastschrifteinzug",
						},
						"credit_debit_indicator": "DBIT",
						"status":                 "BOOK",
						"booking_date":           "2026-04-17",
						"value_date":             "2026-04-17",
					},
					{
						"entry_reference": "000020000000002",
						"transaction_amount": map[string]string{
							"currency": "EUR",
							"amount":   "100.0",
						},
						"debtor": map[string]string{
							"name": "Jane Doe",
						},
						"debtor_account": map[string]string{
							"iban": "DE77112233445566778899", // Anonymized Counterparty IBAN
						},
						"creditor_account": map[string]string{
							"iban": "DE88000000000000000000", // Anonymized User IBAN
						},
						"bank_transaction_code": map[string]string{
							"description": "Gutschrift",
						},
						"credit_debit_indicator": "CRDT",
						"status":                 "BOOK",
						"booking_date":           "2026-04-18",
						"value_date":             "2026-04-18",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/accounts/"+accID+"/balances" {
			w.WriteHeader(http.StatusNotFound) // Simulate balance failure to test robustness
			return
		}
	}

	adapter, server := setupTestAdapter(t, handler)
	defer server.Close()

	txns, balance, err := adapter.FetchTransactions(ctx, userID, accID, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, balance)
	assert.Len(t, txns, 2)

	// Verify Debit Mapping
	assert.Equal(t, -50.0, txns[0].Amount)
	assert.Equal(t, "Generic Insurance Corp", txns[0].CounterpartyName)
	assert.Equal(t, "DE99123456789012345678", txns[0].CounterpartyIban)
	assert.Equal(t, "Lastschrifteinzug", txns[0].BankTransactionCode)

	// Verify Credit Mapping
	assert.Equal(t, 100.0, txns[1].Amount)
	assert.Equal(t, "Jane Doe", txns[1].CounterpartyName)
	assert.Equal(t, "DE77112233445566778899", txns[1].CounterpartyIban)
	assert.Equal(t, "Gutschrift", txns[1].BankTransactionCode)
}
