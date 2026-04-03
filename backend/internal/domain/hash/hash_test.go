package hash_test

import (
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/hash"
)

// ── ForTransaction ──────────────────────────────────────────────────────────

func TestForTransaction_IsDeterministic(t *testing.T) {
	txn := entity.Transaction{
		BookingDate: time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		ValutaDate:  time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		Description: "Lastschrift Netflix",
		Amount:      -15.99,
		Currency:    "EUR",
		Reference:   "MANDATE-001",
	}

	h1 := hash.ForTransaction("DE2354586224568642550", txn)
	h2 := hash.ForTransaction("DE2354586224568642550", txn)

	if h1 != h2 {
		t.Errorf("hash is not deterministic: %q vs %q", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("expected 64 hex chars (SHA-256), got %d", len(h1))
	}
}

func TestForTransaction_DifferentAmountProducesDifferentHash(t *testing.T) {
	base := entity.Transaction{
		BookingDate: time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		ValutaDate:  time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		Description: "Lastschrift Netflix",
		Currency:    "EUR",
	}
	t1 := base
	t1.Amount = -15.99
	t2 := base
	t2.Amount = -17.99

	if hash.ForTransaction("DE01", t1) == hash.ForTransaction("DE01", t2) {
		t.Error("different amounts must produce different hashes")
	}
}

func TestForTransaction_DifferentIBANProducesDifferentHash(t *testing.T) {
	txn := entity.Transaction{
		BookingDate: time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		ValutaDate:  time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
		Description: "Überweisung",
		Amount:      100.00,
		Currency:    "EUR",
	}

	h1 := hash.ForTransaction("DE2354586224568642550", txn)
	h2 := hash.ForTransaction("DE99999999999999999999", txn)

	if h1 == h2 {
		t.Error("different IBANs must produce different hashes")
	}
}

// ── ForBankStatement ─────────────────────────────────────────────────────────

func TestForBankStatement_IsDeterministic(t *testing.T) {
	stmt := entity.BankStatement{
		IBAN:        "DE2354586224568642550",
		StatementNo: 3,
		OldBalance:  2000.00,
		NewBalance:  3503.82,
		Currency:    "EUR",
	}

	h1 := hash.ForBankStatement(stmt)
	h2 := hash.ForBankStatement(stmt)

	if h1 != h2 {
		t.Errorf("hash is not deterministic: %q vs %q", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("expected 64 hex chars (SHA-256), got %d", len(h1))
	}
}

func TestForBankStatement_DifferentBalanceProducesDifferentHash(t *testing.T) {
	stmt1 := entity.BankStatement{IBAN: "DE01", StatementNo: 1, OldBalance: 1000, NewBalance: 1500, Currency: "EUR"}
	stmt2 := stmt1
	stmt2.NewBalance = 9999

	if hash.ForBankStatement(stmt1) == hash.ForBankStatement(stmt2) {
		t.Error("different new balance must produce different hashes")
	}
}

// ── Stamp ────────────────────────────────────────────────────────────────────

func TestStamp_FillsHashesOnStatementAndTransactions(t *testing.T) {
	stmt := entity.BankStatement{
		IBAN:        "DE2354586224568642550",
		StatementNo: 3,
		OldBalance:  2000.00,
		NewBalance:  3503.82,
		Currency:    "EUR",
		Transactions: []entity.Transaction{
			{
				BookingDate: time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
				ValutaDate:  time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
				Description: "Gutschrift",
				Amount:      500.00,
				Currency:    "EUR",
			},
		},
	}

	hash.Stamp(&stmt)

	if stmt.ContentHash == "" {
		t.Error("Stamp must set ContentHash on the statement")
	}
	if stmt.Transactions[0].ContentHash == "" {
		t.Error("Stamp must set ContentHash on each transaction")
	}
}

func TestStamp_IsIdempotent(t *testing.T) {
	stmt := entity.BankStatement{
		IBAN:        "DE2354586224568642550",
		StatementNo: 1,
		OldBalance:  0,
		NewBalance:  1000,
		Currency:    "EUR",
	}

	hash.Stamp(&stmt)
	h1 := stmt.ContentHash
	hash.Stamp(&stmt)
	h2 := stmt.ContentHash

	if h1 != h2 {
		t.Error("Stamp must be idempotent")
	}
}
