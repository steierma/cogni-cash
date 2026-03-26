// Package amazonvisa provides a BankStatementParser adapter for the Amazon
// Visa "Umsätze" XLS export (binary BIFF8 format).
//
// File structure (1-indexed rows as seen in Excel, 0-indexed in code):
//
//	Row  1  (0): empty
//	Row  2  (1): "Amazon Visa - Umsätze"
//	Row  3  (2): empty
//	Row  4  (3): "Datum der Belastung:" | "<DD.MM.YYYY, HH:MM Uhr>"
//	Row  5  (4): "Karteninhaber:"       | "<name>"
//	Row  6  (5): "Referenzkonto:"       | "<IBAN with spaces>"
//	Row  7  (6): "Zeitraum der Bewegung:" | "<DD.MM.YYYY - DD.MM.YYYY>"
//	Row  8  (7): "Kreditkartenlimit:"   | "<amount €>"
//	Row  9  (8): "Verbraucht:"          | "<amount €>"
//	Row 10  (9): empty
//	Row 11 (10): column headers: Datum | Zeit | Karte | Beschreibung | Umsatzkategorie | Betrag | Punkte
//	Row 12 (11): empty
//	Row 13+(12+): transaction rows
//
// Transaction columns (0-based):
//
//	0  Datum          "DD.MM.YYYY "
//	1  Zeit           "HH:MM Uhr"
//	2  Karte          "************2235"
//	3  Beschreibung   merchant name
//	4  Umsatzkategorie  category (main)
//	6  Betrag         "-47,54 €" / "+1.994,14 €"
//	7  Punkte         "+47" / "0"
package amazonvisa

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/extrame/xls"

	"cogni-cash/internal/domain/entity"
)

// ErrFormatMismatch indicates this file is not for this parser
var ErrFormatMismatch = errors.New("amazonvisa parser: file content does not match format")

// Parser implements port.BankStatementParser for the Amazon Visa XLS export.
type Parser struct {
	logger *slog.Logger // Structured logger
}

// NewParser returns a new Amazon Visa XLS Parser.
func NewParser(logger *slog.Logger) *Parser { return &Parser{logger: logger} }

// Parse opens the binary XLS at filePath and returns a BankStatement.
func (p *Parser) Parse(_ context.Context, filePath string) (entity.BankStatement, error) {
	p.logger.Info("Parsing Amazon Visa XLS", "filePath", filePath)
	// Verify the file exists before handing it to the library.
	if _, err := os.Stat(filePath); err != nil {
		p.logger.Error("Failed to stat Amazon Visa XLS", "filePath", filePath, "error", err)
		return entity.BankStatement{}, fmt.Errorf("amazonvisa parser: stat: %w", err)
	}

	wb, err := xls.Open(filePath, "utf-8")
	if err != nil {
		// If it cannot be opened as an XLS, it's definitely not the right format for this parser
		p.logger.Debug("Failed to open file as XLS", "filePath", filePath, "error", err)
		return entity.BankStatement{}, ErrFormatMismatch
	}

	sheet := wb.GetSheet(0)
	if sheet == nil {
		return entity.BankStatement{}, ErrFormatMismatch
	}

	// 1. THE SNIFF TEST: Check if this is an Amazon Visa XLS
	// Row 2 (index 1) usually contains "Amazon Visa - Umsätze"
	titleRow := sheet.Row(1)
	titleText := cellStr(titleRow, 0)
	if !strings.Contains(titleText, "Amazon") && !strings.Contains(titleText, "Visa") {
		return entity.BankStatement{}, ErrFormatMismatch
	}

	stmt := entity.BankStatement{
		SourceFile:    filePath,
		Currency:      "EUR",
		BIC:           "", // Visa card — no BIC
		StatementType: entity.StatementTypeCreditCard,
	}

	// ── parse metadata rows (0–8) ─────────────────────────────────────────────
	for rowIdx := 0; rowIdx <= 8; rowIdx++ {
		row := sheet.Row(rowIdx)
		if row == nil {
			continue
		}
		key := strings.TrimSpace(cellStr(row, 0))
		val := strings.TrimSpace(cellStr(row, 1))

		switch {
		case strings.HasPrefix(key, "Karteninhaber"):
			stmt.AccountHolder = sanitize(val)
		case strings.HasPrefix(key, "Referenzkonto"):
			stmt.IBAN = strings.ReplaceAll(sanitize(val), " ", "")
		case strings.HasPrefix(key, "Zeitraum"):
			// Parse the end date as StatementDate: "01.01.2025 - 31.12.2025"
			// The end date may be absent on open-ended exports ("12.03.2024 - ").
			parts := strings.SplitN(sanitize(val), " - ", 2)
			if len(parts) == 2 {
				if t, err := time.Parse("02.01.2006", strings.TrimSpace(parts[1])); err == nil {
					stmt.StatementDate = t
				}
			}
		case strings.HasPrefix(key, "Verbraucht"):
			// "Verbraucht" is the total spent — use as NewBalance (negative by convention)
			if amt, err := parseEuroAmount(sanitize(val)); err == nil {
				stmt.NewBalance = -amt // credit card spend is a liability
			}
		}
	}

	var newestDate time.Time

	// ── parse transaction rows (row 12 onward, index 12) ──────────────────────
	// Row 10 (index 10) is the column-header row; row 11 (index 11) is blank.
	maxRow := int(sheet.MaxRow)
	for rowIdx := 12; rowIdx <= maxRow; rowIdx++ {
		row := sheet.Row(rowIdx)
		if row == nil {
			continue
		}

		dateStr := strings.TrimSpace(cellStr(row, 0))
		if dateStr == "" {
			continue
		}

		// Pass the parent statement's type down to the transaction
		tx, err := parseTransactionRow(row, stmt.StatementType)
		if err != nil {
			continue // skip malformed rows silently
		}
		stmt.Transactions = append(stmt.Transactions, tx)

		// Track the newest date for fallback StatementDate
		if tx.BookingDate.After(newestDate) {
			newestDate = tx.BookingDate
		}
	}

	// Ensure we have a valid statement date for the global validation
	if stmt.StatementDate.IsZero() && !newestDate.IsZero() {
		stmt.StatementDate = newestDate
	}

	// Derive OldBalance = NewBalance − sum(amounts).
	// For a credit card: credits (payments) are positive, charges are negative.
	var total float64
	for _, tx := range stmt.Transactions {
		total += tx.Amount
	}
	stmt.OldBalance = stmt.NewBalance - total

	// Validation is now handled globally by BankStatementService calling stmt.IsValid()
	p.logger.Info("Successfully parsed Amazon Visa XLS", "filePath", filePath)
	return stmt, nil
}

func parseTransactionRow(row *xls.Row, stmtType entity.StatementType) (entity.Transaction, error) {
	dateStr := strings.TrimSpace(cellStr(row, 0))
	bookingDate, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("amazonvisa: parse date %q: %w", dateStr, err)
	}

	karte := sanitize(cellStr(row, 2))
	description := sanitize(cellStr(row, 3))
	amountStr := strings.TrimSpace(cellStr(row, 6))

	amount, err := parseEuroAmount(amountStr)
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("amazonvisa: parse amount %q: %w", amountStr, err)
	}

	txType := entity.TransactionTypeCredit
	if amount < 0 {
		txType = entity.TransactionTypeDebit
	}

	return entity.Transaction{
		BookingDate:   bookingDate,
		ValutaDate:    bookingDate,
		Description:   description,
		Amount:        amount,
		Currency:      "EUR",
		Type:          txType,
		CategoryID:    nil,
		Reference:     karte,
		StatementType: stmtType, // Explicitly capture the type here
	}, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func cellStr(row *xls.Row, col int) string {
	if row == nil {
		return ""
	}
	return row.Col(col)
}

// parseEuroAmount parses German-formatted Euro strings like "-47,54 €" or "+1.994,14 €".
// Returns the float64 value (negative for debits, positive for credits).
func parseEuroAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	// Remove € sign and surrounding spaces.
	s = strings.ReplaceAll(s, "€", "")
	s = strings.TrimSpace(s)
	// Remove thousands separator (German: full stop).
	s = strings.ReplaceAll(s, ".", "")
	// Replace decimal comma with full stop.
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

// sanitize trims whitespace and removes null bytes that the extrame/xls library
// occasionally embeds in strings from older BIFF8 files.  PostgreSQL rejects
// any string containing \x00, which would otherwise cause a 422 on import.
func sanitize(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(s)
}

// xlsCategoryAliases maps the mangled category strings that the extrame/xls
// library emits (it silently drops some non-ASCII bytes from BIFF8 strings)
// back to the canonical names seeded in the categories table.
var xlsCategoryAliases = map[string]string{
	"Handel und Geschfte":                   "Handel und Geschäfte",
	"berweisungen, Bankkosten und Darlehen": "Überweisungen, Bankkosten und Darlehen",
}

// normalizeCategory maps a raw Umsatzkategorie string from the XLS to the
// canonical category name stored in the categories table.  Known encoding
// artefacts (dropped umlauts) are repaired via the alias table above.
func normalizeCategory(raw string) string {
	if canonical, ok := xlsCategoryAliases[raw]; ok {
		return canonical
	}
	return raw
}
