package ingcsv_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"

	"cogni-cash/internal/adapter/parser/bank_statement/ingcsv"
)

func csvPath() string {
	_, file, _, _ := runtime.Caller(0)
	// file is at internal/adapter/parser/ingcsv/parser_test.go
	// walk up 5 levels to reach the backend/ root, then into balance/
	root := filepath.Dir(file)
	for range [5]struct{}{} {
		root = filepath.Dir(root)
	}
	return filepath.Join(root, "balance", "Umsatzanzeige_02_2026.csv")
}

func TestINGCSVParser_ParseRealCSV(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	p := ingcsv.NewParser(logger)

	fileBytes, err := os.ReadFile(csvPath())
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	stmt, err := p.Parse(context.Background(), uuid.Nil, fileBytes)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	t.Logf("Account Holder : %s", stmt.AccountHolder)
	t.Logf("IBAN           : %s", stmt.IBAN)
	t.Logf("New Balance    : %.2f %s", stmt.NewBalance, stmt.Currency)
	t.Logf("Old Balance    : %.2f %s", stmt.OldBalance, stmt.Currency)
	t.Logf("Transactions   : %d", len(stmt.Transactions))
	for i, tx := range stmt.Transactions {
		t.Logf("  [%02d] %s  %-60s  %+9.2f EUR  (%s) ref=%q",
			i+1,
			tx.BookingDate.Format("02.01.2006"),
			tx.Description,
			tx.Amount,
			tx.Type,
			tx.Reference)
	}

	// ---- assertions based on known CSV content --------------------------------

	if stmt.IBAN != "DE89370400440532013000" {
		t.Errorf("expected IBAN DE89370400440532013000, got %q", stmt.IBAN)
	}
	if stmt.AccountHolder != "Erika Mustermann" {
		t.Errorf("expected account holder Erika Mustermann, got %q", stmt.AccountHolder)
	}
	if stmt.NewBalance != -352.76 {
		t.Errorf("expected new balance -352.76, got %.2f", stmt.NewBalance)
	}
	if len(stmt.Transactions) != 58 {
		t.Errorf("expected 58 transactions, got %d", len(stmt.Transactions))
	}

	// First transaction in the file is the most recent: salary on 27.02.2026
	first := stmt.Transactions[0]
	if first.Amount != 4752.01 {
		t.Errorf("expected first transaction amount 4752.01 (salary), got %.2f", first.Amount)
	}
	if first.Type != "credit" {
		t.Errorf("expected first transaction type credit, got %s", first.Type)
	}
	last := stmt.Transactions[len(stmt.Transactions)-1]
	if last.Amount != 1300.00 {
		t.Errorf("expected last transaction amount 1300.00, got %.2f", last.Amount)
	}
}
