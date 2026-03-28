// Package ingcsv provides a BankStatementParser adapter for the ING-DiBa
// "Umsatzanzeige" CSV export (Girokonto).
//
// File characteristics:
//   - Encoding:  ISO-8859-1
//   - Separator: semicolon (;)
//   - Structure:
//     Line 1:  "Umsatzanzeige;Datei erstellt am: DD.MM.YYYY HH:MM"
//     Line 2:  blank
//     Line 3:  "IBAN;<iban value>"
//     Line 4:  "Kontoname;Girokonto"
//     Line 5:  "Bank;ING"
//     Line 6:  "Kunde;<name>"
//     Line 7:  "Zeitraum;<from> - <to>"
//     Line 8:  "Saldo;<amount>;<currency>"
//     Line 9:  blank
//     Line 10: "Sortierung;..."
//     Line 11: info paragraph
//     Line 12: blank
//     Line 13: column header row
//     Line 14+: transaction rows
//
// Transaction columns (0-based):
//
//	0  Buchung               DD.MM.YYYY
//	1  Wertstellungsdatum    DD.MM.YYYY
//	2  Auftraggeber/Empfänger
//	3  Buchungstext
//	4  Verwendungszweck
//	5  Saldo                 German float (running balance)
//	6  Währung               EUR
//	7  Betrag                German float (positive = credit, negative = debit)
//	8  Währung               EUR
package ingcsv

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"cogni-cash/internal/domain/entity"
	"log/slog"
)

// ErrFormatMismatch indicates this file is not for this parser
var ErrFormatMismatch = errors.New("ingcsv parser: file content does not match format")

// Parser implements port.BankStatementParser for the ING CSV export.
type Parser struct {
	Logger *slog.Logger // Structured logger
}

// NewParser returns a new ING CSV Parser.
func NewParser(logger *slog.Logger) *Parser { return &Parser{Logger: logger} }

// Parse opens the ISO-8859-1 CSV at filePath and returns a BankStatement.
func (p *Parser) Parse(_ context.Context, filePath string) (entity.BankStatement, error) {
	p.Logger.Info("Parsing ING CSV", "filePath", filePath)
	f, err := os.Open(filePath)
	if err != nil {
		p.Logger.Error("Failed to open ING CSV", "filePath", filePath, "error", err)
		return entity.BankStatement{}, fmt.Errorf("ingcsv parser: open: %w", err)
	}
	defer f.Close()

	// Decode ISO-8859-1 → UTF-8 on the fly.
	utf8Reader := transform.NewReader(f, charmap.ISO8859_1.NewDecoder())

	r := csv.NewReader(utf8Reader)
	r.Comma = ';'
	r.LazyQuotes = true
	r.FieldsPerRecord = -1 // variable — header rows have fewer columns

	// Read all records at once; the file is small.
	records, err := r.ReadAll()
	if err != nil {
		return entity.BankStatement{}, fmt.Errorf("ingcsv parser: read csv: %w", err)
	}

	// 1. THE SNIFF TEST: Check if this is an ING CSV
	// ING CSVs strictly start with "Umsatzanzeige" in the first cell.
	if len(records) == 0 || len(records[0]) == 0 || !strings.Contains(records[0][0], "Umsatzanzeige") {
		return entity.BankStatement{}, ErrFormatMismatch
	}

	stmt := entity.BankStatement{
		SourceFile: filePath,
		Currency:   "EUR",
	}

	// ---- parse metadata rows ------------------------------------------------
	for _, row := range records {
		if len(row) < 2 {
			continue
		}
		key := strings.TrimSpace(row[0])
		val := strings.TrimSpace(row[1])

		switch key {
		case "IBAN":
			stmt.IBAN = strings.ReplaceAll(val, " ", "")
		case "Kunde":
			// "Max Mustermann, Max Mustermann" → first name only
			parts := strings.SplitN(val, ",", 2)
			stmt.AccountHolder = strings.TrimSpace(parts[0])
		case "Zeitraum":
			// "01.02.2026 - 28.02.2026"
			// PeriodLabel removed
		case "Saldo":
			if len(row) >= 3 {
				stmt.NewBalance, _ = parseGermanFloat(val)
				stmt.Currency = strings.TrimSpace(row[2])
			}
		case "Buchung":
			// This is the column-header row — everything after here is transactions.
			break
		}
	}

	// BIC is not in the CSV; set a known constant for ING Germany.
	stmt.BIC = "INGDDEFFXXX"

	// Derive StatementDate from the newest transaction, as the CSV lacks a strict statement date
	var newestDate time.Time

	// ---- find the transaction data rows -------------------------------------
	// Transactions start after the header row whose first field is "Buchung".
	txStart := -1
	for i, row := range records {
		if len(row) >= 9 && strings.TrimSpace(row[0]) == "Buchung" {
			txStart = i + 1
			break
		}
	}

	if txStart != -1 {
		for _, row := range records[txStart:] {
			if len(row) < 9 {
				continue
			}
			tx, err := parseRow(row)
			if err != nil {
				continue // skip malformed rows silently
			}
			stmt.Transactions = append(stmt.Transactions, tx)

			// Track the newest date for the StatementDate
			if tx.BookingDate.After(newestDate) {
				newestDate = tx.BookingDate
			}
		}
	}

	// Ensure we have a valid statement date for the global validation
	if !newestDate.IsZero() {
		stmt.StatementDate = newestDate
	}

	// Derive OldBalance = NewBalance − sum(amounts).
	var total float64
	for _, tx := range stmt.Transactions {
		total += tx.Amount
	}
	stmt.OldBalance = stmt.NewBalance - total

	// Validation is now handled globally by BankStatementService calling stmt.IsValid()
	p.Logger.Info("Successfully parsed ING CSV", "filePath", filePath)
	return stmt, nil
}

// parseRow converts a single 9-field transaction CSV row into a Transaction.
//
// Col 0: Buchung               → BookingDate
// Col 1: Wertstellungsdatum    → ValutaDate
// Col 2: Auftraggeber/Empfänger → part of Description
// Col 3: Buchungstext           → part of Description
// Col 4: Verwendungszweck       → Reference (kept separate for AI categorisation)
// Col 5: Saldo                  (running balance — not stored)
// Col 6: Währung                → Currency
// Col 7: Betrag                 → Amount
// Col 8: Währung                (duplicate — ignored)
func parseRow(row []string) (entity.Transaction, error) {
	bookingDate, err := parseGermanDate(row[0])
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("booking date: %w", err)
	}
	valutaDate, err := parseGermanDate(row[1])
	if err != nil {
		valutaDate = bookingDate
	}

	counterparty := strings.TrimSpace(row[2])
	bookingType := strings.TrimSpace(row[3])
	reference := strings.TrimSpace(row[4]) // Verwendungszweck → Reference field
	currency := strings.TrimSpace(row[6])
	if currency == "" {
		currency = "EUR"
	}

	amount, err := parseGermanFloat(row[7])
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("amount: %w", err)
	}

	// Description = Auftraggeber/Empfänger + Buchungstext only.
	// Verwendungszweck is intentionally kept in Reference so the AI layer can
	// use it independently for categorisation without parsing it out of a
	// concatenated string.
	descParts := []string{counterparty}
	if bookingType != "" && bookingType != counterparty {
		descParts = append(descParts, bookingType)
	}
	description := strings.Join(descParts, " | ")

	txType := entity.TransactionTypeCredit
	if amount < 0 {
		txType = entity.TransactionTypeDebit
	}

	return entity.Transaction{
		BookingDate: bookingDate,
		ValutaDate:  valutaDate,
		Description: description,
		Amount:      amount,
		Currency:    currency,
		Type:        txType,
		Reference:   reference, // Verwendungszweck
		CategoryID:  nil,       // Auto-categorizer will catch this since we migrated to UUID mappings
	}, nil
}

// ---- helpers ---------------------------------------------------------------

func parseGermanFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

func parseGermanDate(s string) (time.Time, error) {
	return time.Parse("02.01.2006", strings.TrimSpace(s))
}
