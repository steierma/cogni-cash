package vw

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

var ErrFormatMismatch = errors.New("vw parser: file content does not match format")

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(ctx context.Context, _ uuid.UUID, fileBytes []byte) (entity.BankStatement, error) {
	rawText, err := extractPDFText(fileBytes)
	if err != nil {
		return entity.BankStatement{}, fmt.Errorf("vw parser: failed to read pdf: %w", err)
	}

	upperText := strings.ToUpper(rawText)
	if !strings.Contains(upperText, "VOLKSWAGEN") && !strings.Contains(upperText, "VW BANK") {
		return entity.BankStatement{}, ErrFormatMismatch
	}

	stmt := entity.BankStatement{
		StatementType: entity.StatementTypeGiro,
	}

	if strings.Contains(rawText, "Plus Konto online") {
		stmt.StatementType = entity.StatementTypeExtraAccount
	}

	ibanRegex := regexp.MustCompile(`IBAN:\s*([A-Z0-9]{22})`)
	if match := ibanRegex.FindStringSubmatch(rawText); len(match) > 1 {
		stmt.IBAN = match[1]
	}

	dateRegex := regexp.MustCompile(`Erstellungsdatum:\s*(\d{2}\.\d{2}\.\d{4})`)
	if match := dateRegex.FindStringSubmatch(rawText); len(match) > 1 {
		stmt.StatementDate, _ = time.Parse("02.01.2006", match[1])
	}

	newBalanceRegex := regexp.MustCompile(`(?i)Neuer Kontostand in EUR:\s*([\d\.]+,\d{2})`)
	if match := newBalanceRegex.FindStringSubmatch(rawText); len(match) > 1 {
		stmt.NewBalance, _ = parseGermanFloat(match[1])
	}

	stmt.Transactions = parseTransactions(rawText)

	return stmt, nil
}

func parseTransactions(text string) []entity.Transaction {
	start := strings.Index(text, "Umsatzinformationen")
	if start != -1 {
		text = text[start:]
	}

	end := strings.Index(text, "Neuer Kontostand")
	if end == -1 {
		end = strings.Index(text, "Freistellungs-")
	}
	if end != -1 {
		text = text[:end]
	}

	normalizedText := strings.ReplaceAll(text, "\n", " ")
	tokens := strings.Fields(normalizedText)

	dateRegex := regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)
	amountRegex := regexp.MustCompile(`^-?\d{1,3}(?:\.\d{3})*,\d{2}$`)

	var txns []entity.Transaction
	var currentChunk []string
	var datesInChunk []time.Time

	for _, tok := range tokens {
		if dateRegex.MatchString(tok) {
			d, _ := time.Parse("02.01.2006", tok)
			datesInChunk = append(datesInChunk, d)
			currentChunk = append(currentChunk, tok)
		} else if amountRegex.MatchString(tok) {
			// A true transaction amount is almost always immediately preceded by the Valuta Date.
			isTxAmount := false
			if len(datesInChunk) >= 2 && len(currentChunk) > 0 {
				lastTok := currentChunk[len(currentChunk)-1]
				if dateRegex.MatchString(lastTok) {
					isTxAmount = true
				}
			}

			if isTxAmount {
				amt, _ := parseGermanFloat(tok)

				// Take the last two dates encountered before the amount
				bookingDate := datesInChunk[len(datesInChunk)-2]
				valutaDate := datesInChunk[len(datesInChunk)-1]

				currentChunk = append(currentChunk, tok)
				desc := extractMainDesc(currentChunk)

				txns = append(txns, entity.Transaction{
					BookingDate: bookingDate,
					ValutaDate:  valutaDate,
					Description: desc,
					Amount:      amt,
				})

				// Reset state for the next transaction block
				currentChunk = []string{}
				datesInChunk = []time.Time{}
			} else {
				// Not a transaction amount (e.g. "15,00" embedded in a detail note)
				currentChunk = append(currentChunk, tok)
			}
		} else {
			currentChunk = append(currentChunk, tok)
		}
	}

	return txns
}

// extractMainDesc scans the text chunk belonging to a single transaction and isolates the core action.
func extractMainDesc(tokens []string) string {
	text := strings.Join(tokens, " ")
	keywords := []string{
		"Telebanking Belastung",
		"Telebanking Gutschrift",
		"Gutschrift",
		"Habenzinsen",
		"Sollzinsen",
		"Lastschrift",
		"Überweisung",
		"Dauerauftrag",
		"Entgelt",
		"Scheck",
	}

	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return kw
		}
	}

	// Fallback: Return a sanitized version of the text if no keyword matches
	var clean []string
	dateRegex := regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)
	for _, t := range tokens {
		if !dateRegex.MatchString(t) && !isNumericID(t) {
			clean = append(clean, t)
		}
	}
	return strings.TrimSpace(strings.Join(clean, " "))
}

func isNumericID(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil && len(s) <= 3
}

func parseGermanFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

func extractPDFText(fileBytes []byte) (string, error) {
	readerAt := bytes.NewReader(fileBytes)
	r, err := pdf.NewReader(readerAt, int64(len(fileBytes)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}
