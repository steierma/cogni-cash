package vwcsv

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

var ErrFormatMismatch = errors.New("vwcsv parser: file content does not match format")

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(_ context.Context, _ uuid.UUID, fileBytes []byte) (entity.BankStatement, error) {
	if !p.isCSV(fileBytes) {
		return entity.BankStatement{}, ErrFormatMismatch
	}

	r := csv.NewReader(bytes.NewReader(fileBytes))
	r.Comma = ';'
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return entity.BankStatement{}, fmt.Errorf("vwcsv parser: failed to read csv: %w", err)
	}

	if len(records) == 0 {
		return entity.BankStatement{}, ErrFormatMismatch
	}

	stmt := entity.BankStatement{
		StatementType: entity.StatementTypeGiro,
		Currency:      "EUR",
	}

	var newestDate time.Time

	for i, row := range records {
		if len(row) == 0 {
			continue
		}

		// Metadata sniffing
		for _, cell := range row {
			if strings.Contains(cell, "Plus Konto online Nr.") {
				stmt.StatementType = entity.StatementTypeExtraAccount
				parts := strings.Split(cell, "Nr.")
				if len(parts) > 1 {
					stmt.IBAN = strings.TrimSpace(parts[1])
				}
			}
		}

		if len(row) > 1 && row[0] == "Saldo (EUR)" {
			stmt.NewBalance, _ = parseGermanFloat(row[1])
		} else if len(row) > 1 && row[0] == "Zeitraum" {
			parts := strings.Split(row[1], "-")
			if len(parts) > 1 {
				stmt.StatementDate, _ = time.Parse("02.01.2006", strings.TrimSpace(parts[1]))
			}
		}

		// Header detection for transactions
		if row[0] == "Nr." && len(row) >= 12 && row[1] == "Buchungsdatum" {
			for j := i + 1; j < len(records); j++ {
				txRow := records[j]
				if len(txRow) < 12 || txRow[1] == "" {
					continue
				}

				bookingDate, _ := time.Parse("02.01.2006", txRow[1])
				valutaDate, _ := time.Parse("02.01.2006", txRow[9])
				if valutaDate.IsZero() {
					valutaDate = bookingDate
				}

				var amount float64
				if txRow[10] != "" {
					soll, _ := parseGermanFloat(txRow[10])
					amount = -soll
				} else if txRow[11] != "" {
					haben, _ := parseGermanFloat(txRow[11])
					amount = haben
				}

				tx := entity.Transaction{
					BookingDate: bookingDate,
					ValutaDate:  valutaDate,
					Description: strings.TrimSpace(txRow[2]),
					Reference:   strings.TrimSpace(txRow[3]),
					Amount:      amount,
					Currency:    "EUR",
					Type:        entity.TransactionTypeCredit,
				}
				if tx.Description == "" {
					tx.Description = tx.Reference
					tx.Reference = ""
				}

				if amount < 0 {
					tx.Type = entity.TransactionTypeDebit
				}

				stmt.Transactions = append(stmt.Transactions, tx)

				if bookingDate.After(newestDate) {
					newestDate = bookingDate
				}
			}
			break
		}
	}

	if stmt.StatementDate.IsZero() && !newestDate.IsZero() {
		stmt.StatementDate = newestDate
	}

	if stmt.IBAN == "" {
		for _, row := range records {
			for _, cell := range row {
				if strings.Contains(cell, "Nr. ") {
					parts := strings.Split(cell, "Nr.")
					if len(parts) > 1 {
						stmt.IBAN = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}
	}

	return stmt, nil
}

func (p *Parser) isCSV(fileBytes []byte) bool {
	head := string(fileBytes[:min(500, len(fileBytes))])
	return strings.Contains(head, "Kontoinhaber;") || strings.Contains(head, "Plus Konto online Nr.")
}

func min(a, b int) int {
	if a < b { return a }; return b
}

func parseGermanFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}
