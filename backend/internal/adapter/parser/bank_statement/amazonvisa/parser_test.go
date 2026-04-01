package amazonvisa

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// xlsPath resolves the path to the real test fixture relative to this file.
func xlsPath() string {
	_, file, _, _ := runtime.Caller(0)
	// file is at internal/adapter/parser/amazonvisa/parser_test.go
	// walk up 5 levels to reach the backend/ root, then into balance/
	root := filepath.Dir(file)
	for range [5]struct{}{} {
		root = filepath.Dir(root)
	}
	return filepath.Join(root, "balance", "Amazon_Visa_25_11_2025_bis_13_03_2026.xls")
}

func TestAmazonVisaParser_ParseRealXLS(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	p := NewParser(logger)
	stmt, err := p.Parse(context.Background(), uuid.Nil, xlsPath())
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	t.Logf("Account Holder : %s", stmt.AccountHolder)
	t.Logf("IBAN           : %s", stmt.IBAN)
	t.Logf("Statement Date : %s", stmt.StatementDate.Format("2006-01-02"))
	t.Logf("New Balance    : %.2f %s", stmt.NewBalance, stmt.Currency)
	t.Logf("Old Balance    : %.2f %s", stmt.OldBalance, stmt.Currency)
	t.Logf("Transactions   : %d", len(stmt.Transactions))
	for i, tx := range stmt.Transactions {
		t.Logf("  [%02d] %s  %-50s  %+.2f %s  (%s)",
			i+1, tx.BookingDate.Format("02.01.2006"), tx.Description, tx.Amount, tx.Currency, tx.Type)
	}

	// ── structural assertions ─────────────────────────────────────────────────

	if stmt.AccountHolder != "Max Mustermann" {
		t.Errorf("account holder: got %q, want %q", stmt.AccountHolder, "Max Mustermann")
	}
	if stmt.IBAN != "DE89370400440532013000" {
		t.Errorf("IBAN: got %q, want %q", stmt.IBAN, "DE89370400440532013000")
	}
	// StatementDate is derived from the Datum der Belastung / Zeitraum fields
	// in the actual fixture on disk.
	wantDate := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	if !stmt.StatementDate.Equal(wantDate) {
		t.Errorf("statement date: got %v, want %v", stmt.StatementDate, wantDate)
	}
	if stmt.Currency != "EUR" {
		t.Errorf("currency: got %q, want %q", stmt.Currency, "EUR")
	}
	if len(stmt.Transactions) != 31 {
		t.Errorf("transaction count: got %d, want 31", len(stmt.Transactions))
	}

	// First transaction in the file (row index 0 = earliest purchase):
	// 25.11.2025  Lebensmittelmarkt Muster  -87.49
	first := stmt.Transactions[0]
	if first.BookingDate.Format("02.01.2006") != "25.11.2025" {
		t.Errorf("first tx date: got %q, want %q", first.BookingDate.Format("02.01.2006"), "25.11.2025")
	}
	if first.Amount != -87.49 {
		t.Errorf("first tx amount: got %.2f, want -87.49", first.Amount)
	}

	// Last transaction must be the Girokonto payment credit of +2500.00.
	last := stmt.Transactions[len(stmt.Transactions)-1]
	if last.BookingDate.Format("02.01.2006") != "15.03.2026" {
		t.Errorf("payment tx date: got %q, want %q", last.BookingDate.Format("02.01.2006"), "15.03.2026")
	}
	if last.Amount != 2500.00 {
		t.Errorf("payment tx amount: got %.2f, want +2500.00", last.Amount)
	}
	if last.Type != "credit" {
		t.Errorf("payment tx type: got %q, want %q", last.Type, "credit")
	}

	// All transactions must have a non-empty description, valid date, and no null bytes.
	for i, tx := range stmt.Transactions {
		if tx.Description == "" {
			t.Errorf("tx[%d]: empty description", i)
		}
		if tx.BookingDate.IsZero() {
			t.Errorf("tx[%d]: zero booking date", i)
		}
		if tx.Currency != "EUR" {
			t.Errorf("tx[%d]: currency %q, want EUR", i, tx.Currency)
		}
		// PostgreSQL rejects strings containing null bytes — ensure none slip through.
		if strings.Contains(tx.Description, "\x00") {
			t.Errorf("tx[%d]: description contains null byte: %q", i, tx.Description)
		}
		if strings.Contains(tx.Reference, "\x00") {
			t.Errorf("tx[%d]: reference contains null byte: %q", i, tx.Reference)
		}
	}
}

func TestAmazonVisaParser_MissingFile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	p := NewParser(logger)
	_, err := p.Parse(context.Background(), uuid.Nil, "/nonexistent/file.xls")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
