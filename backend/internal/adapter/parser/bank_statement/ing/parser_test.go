package ing

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
)

// pdfPath returns the absolute path to the test fixture PDF.
func pdfPath() string {
	_, file, _, _ := runtime.Caller(0)
	// File is at internal/adapter/parser/ing/parser_test.go
	// Walk up 4 levels to reach the project root.
	root := filepath.Dir(file) // .../internal/adapter/parser/ing
	for range [5]struct{}{} {
		root = filepath.Dir(root)
	}
	return filepath.Join(root, "balance", "Girokonto_5437817550_Kontoauszug_20260301.pdf")
}

func TestINGParser_ParseRealPDF(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	p := NewParser(logger)

	fileBytes, err := os.ReadFile(pdfPath())
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	stmt, err := p.Parse(context.Background(), uuid.Nil, fileBytes)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	t.Logf("Account Holder : %s", stmt.AccountHolder)
	t.Logf("IBAN           : %s", stmt.IBAN)
	t.Logf("Statement No   : %d", stmt.StatementNo)
	t.Logf("Statement Date : %s", stmt.StatementDate.Format("02.01.2006"))
	t.Logf("Old Balance    : %.2f %s", stmt.OldBalance, stmt.Currency)
	t.Logf("New Balance    : %.2f %s", stmt.NewBalance, stmt.Currency)
	t.Logf("Transactions   : %d", len(stmt.Transactions))

	for i, tx := range stmt.Transactions {
		catStr := "<nil>"
		if tx.CategoryID != nil {
			catStr = tx.CategoryID.String()
		}
		t.Logf("  [%02d] %s  %-55s  %+8.2f EUR  (%s)  ref=%q  cat=%s",
			i+1,
			tx.BookingDate.Format("02.01.2006"),
			tx.Description,
			tx.Amount,
			tx.Type,
			tx.Reference,
			catStr)
	}

	// Assertions based on the known content of the PDF
	if stmt.IBAN != "DE89370400440532013000" {
		t.Errorf("expected IBAN DE89370400440532013000, got %q", stmt.IBAN)
	}
	if stmt.NewBalance != 3503.82 {
		t.Errorf("expected new balance 3503.82, got %.2f", stmt.NewBalance)
	}
	if stmt.OldBalance != 1758.24 {
		t.Errorf("expected old balance 1758.24, got %.2f", stmt.OldBalance)
	}
	if stmt.StatementNo != 2 {
		t.Errorf("expected statement no 2, got %d", stmt.StatementNo)
	}
	if len(stmt.Transactions) == 0 {
		t.Fatal("expected transactions to be parsed, got 0")
	}

	// Verify the first transaction: salary credit on 27.02.2026
	// Salary credit (last transaction)
	last := stmt.Transactions[len(stmt.Transactions)-1]
	if last.Amount != 4752.01 {
		t.Errorf("expected last transaction (salary) to be 4752.01, got %.2f", last.Amount)
	}

	// ING PDF has no source-provided category — field must always be nil.
	for i, tx := range stmt.Transactions {
		if tx.CategoryID != nil {
			t.Errorf("tx[%d]: expected nil category for ING PDF, got %q", i, tx.CategoryID)
		}
	}

	// Every transaction must have a non-empty description and a valid date.
	for i, tx := range stmt.Transactions {
		if tx.Description == "" {
			t.Errorf("tx[%d]: empty description", i)
		}
		if tx.BookingDate.IsZero() {
			t.Errorf("tx[%d]: zero booking date", i)
		}
		// Reference is a string field; it may be empty for some rows but must not panic.
		_ = tx.Reference
	}
}
