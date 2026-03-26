// Package hash provides pure, deterministic content-hash helpers for domain
// entities. No external dependencies — only stdlib crypto/sha256.
//
// These hashes are used as idempotency keys so that re-importing the same
// bank statement or transaction is a no-op rather than creating a duplicate.
package hash

import (
	"crypto/sha256"
	"fmt"

	"cogni-cash/internal/domain/entity"
)

// ForTransaction returns a stable SHA-256 hex string over the fields that
// uniquely identify a transaction:
//
//	IBAN + booking_date + valuta_date + description + amount + currency + reference
//
// The IBAN is passed in separately because Transaction itself doesn't carry it.
func ForTransaction(iban string, t entity.Transaction) string {
	h := sha256.New()
	// Use a pipe-delimited payload — field order is fixed and must never change.
	fmt.Fprintf(h, "%s|%s|%s|%s|%.2f|%s|%s",
		iban,
		t.BookingDate.Format("2006-01-02"),
		t.ValutaDate.Format("2006-01-02"),
		t.Description,
		t.Amount,
		t.Currency,
		t.Reference,
	)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ForBankStatement returns a stable SHA-256 hex string over the fields that
// uniquely identify a bank statement:
//
//	IBAN + statement_no + old_balance + new_balance + currency
//
// Source file name is intentionally excluded so that renaming a file does not
// change the hash of an already-imported statement.
func ForBankStatement(stmt entity.BankStatement) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%d|%.2f|%.2f|%s",
		stmt.IBAN,
		stmt.StatementNo,
		stmt.OldBalance,
		stmt.NewBalance,
		stmt.Currency,
	)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Stamp computes and injects ContentHash values into a BankStatement and all
// of its Transactions in-place. Call this once after parsing, before persisting.
func Stamp(stmt *entity.BankStatement) {
	stmt.ContentHash = ForBankStatement(*stmt)
	for i := range stmt.Transactions {
		stmt.Transactions[i].ContentHash = ForTransaction(stmt.IBAN, stmt.Transactions[i])
	}
}
