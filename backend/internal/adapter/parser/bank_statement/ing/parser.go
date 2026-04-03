package ing

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

	"log/slog"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

// ErrFormatMismatch indicates this file is not for this parser
var ErrFormatMismatch = errors.New("ing parser: file content does not match format")

// ---- compiled regexps ------------------------------------------------------

var (
	reDate          = regexp.MustCompile(`^\d{2}\.\d{2}\.\d{4}$`)
	reAmount        = regexp.MustCompile(`^-?\d{1,3}(?:\.\d{3})*,\d{2}$`)
	reIBAN          = regexp.MustCompile(`DE\d{2}(?:\s?\d{4}){5}`)
	reBIC           = regexp.MustCompile(`[A-Z]{4}DE[A-Z0-9]{2}(?:[A-Z0-9]{3})?`)
	reOldBalance    = regexp.MustCompile(`Alter Saldo\s+([\d.]+,\d{2})\s*Euro`)
	reNewBalance    = regexp.MustCompile(`Neuer Saldo\s+([\d.]+,\d{2})\s*Euro`)
	reStatementDate = regexp.MustCompile(`Datum\s*\n?\s*(\d{2}\.\d{2}\.\d{4})`)
	reStatementNo   = regexp.MustCompile(`Auszugsnummer\s*\n?\s*(\d+)`)
	reAccountNo     = regexp.MustCompile(`Girokonto Nummer (\d+)`)
)

type Parser struct {
	Logger *slog.Logger
}

func NewParser(logger *slog.Logger) *Parser { return &Parser{Logger: logger} }

func (p *Parser) Parse(_ context.Context, _ uuid.UUID, fileBytes []byte) (entity.BankStatement, error) {
	p.Logger.Info("Parsing ING PDF")
	tokens, raw, err := extractTokens(fileBytes)
	if err != nil {
		return entity.BankStatement{}, fmt.Errorf("ing parser: extract text: %w", err)
	}

	// 1. THE SNIFF TEST: Be specific to avoid false positives from other banks
	// mentioning ING in transaction descriptions.
	isING := strings.Contains(raw, "ING-DiBa AG") ||
		(strings.Contains(raw, "ING") && strings.Contains(raw, "60591 Frankfurt"))

	if !isING {
		return entity.BankStatement{}, ErrFormatMismatch
	}

	stmt := entity.BankStatement{
		Currency: "EUR",
	}

	if m := reOldBalance.FindStringSubmatch(raw); len(m) == 2 {
		stmt.OldBalance, _ = parseGermanFloat(m[1])
	}
	if m := reNewBalance.FindStringSubmatch(raw); len(m) == 2 {
		stmt.NewBalance, _ = parseGermanFloat(m[1])
	}
	if m := reStatementDate.FindStringSubmatch(raw); len(m) == 2 {
		stmt.StatementDate, _ = parseGermanDate(m[1])
	}
	if m := reStatementNo.FindStringSubmatch(raw); len(m) == 2 {
		stmt.StatementNo, _ = strconv.Atoi(m[1])
	}
	stmt.AccountHolder = extractAccountHolder(tokens)

	if iban := extractLabeledValue(tokens, "IBAN"); iban != "" {
		stmt.IBAN = strings.ReplaceAll(iban, " ", "")
	} else if m := reIBAN.FindString(raw); m != "" {
		stmt.IBAN = strings.ReplaceAll(m, " ", "")
	}

	stmt.Transactions = parseTransactions(tokens)

	p.Logger.Info("Successfully parsed ING PDF")
	return stmt, nil
}

func extractTokens(fileBytes []byte) ([]string, string, error) {
	readerAt := bytes.NewReader(fileBytes)
	r, err := pdf.NewReader(readerAt, int64(len(fileBytes)))
	if err != nil {
		return nil, "", err
	}

	var tokens []string
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		for _, l := range strings.Split(text, "\n") {
			t := strings.TrimSpace(l)
			if t != "" {
				tokens = append(tokens, t)
			}
		}
	}
	return tokens, strings.Join(tokens, "\n"), nil
}

func extractAccountHolder(tokens []string) string {
	for i, t := range tokens {
		if !strings.HasPrefix(t, "ING-DiBa AG") {
			continue
		}
		for j := i - 1; j >= 0 && j < len(tokens); j-- {
			c := strings.TrimSpace(tokens[j])
			if c == "" {
				continue
			}
			if reDate.MatchString(c) || reAmount.MatchString(c) {
				break
			}
			if strings.HasPrefix(c, "und ") || len(c) <= 2 {
				continue
			}
			if c[0] >= '0' && c[0] <= '9' {
				continue
			}
			skip := false
			for _, kw := range []string{"Sportplatz", "Mandat:", "Referenz:", "Kundennr", "Rechnungsnr"} {
				if strings.Contains(c, kw) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			return c
		}
		break
	}
	return ""
}

func extractLabeledValue(tokens []string, label string) string {
	for i, t := range tokens {
		if t == label && i+1 < len(tokens) {
			return strings.TrimSpace(tokens[i+1])
		}
	}
	return ""
}

func parseTransactions(tokens []string) []entity.Transaction {
	tokens = cutAtFooter(tokens)
	tokens = stripAllPageHeaders(tokens)

	type txRegion struct {
		booking    time.Time
		valuta     time.Time
		descTokens []string
		amountStr  string
		contTokens []string
	}

	var regions []txRegion
	var cur *txRegion
	state := 0 // 0:Idle, 1:Desc, 2:AfterAmount, 3:Continuation

	for _, t := range tokens {
		if isBoilerplate(t) {
			continue
		}
		isDate := reDate.MatchString(t)
		isAmt := reAmount.MatchString(t)

		switch state {
		case 0:
			if isDate {
				d, err := parseGermanDate(t)
				if err == nil {
					cur = &txRegion{booking: d, valuta: d}
					state = 1
				}
			}
		case 1:
			if isAmt {
				cur.amountStr = t
				state = 2
			} else if isDate {
				d, err := parseGermanDate(t)
				if err == nil {
					cur = &txRegion{booking: d, valuta: d}
				}
			} else {
				cur.descTokens = append(cur.descTokens, t)
			}
		case 2:
			if isDate {
				d, _ := parseGermanDate(t)
				cur.valuta = d
				state = 3
			} else {
				cur.contTokens = append(cur.contTokens, t)
				state = 3
			}
		case 3:
			if isDate {
				if cur != nil && cur.amountStr != "" {
					regions = append(regions, *cur)
				}
				d, err := parseGermanDate(t)
				if err == nil {
					cur = &txRegion{booking: d, valuta: d}
					state = 1
				} else {
					cur = nil
					state = 0
				}
			} else {
				cur.contTokens = append(cur.contTokens, t)
			}
		}
	}

	if cur != nil && cur.amountStr != "" {
		regions = append(regions, *cur)
	}

	var txns []entity.Transaction
	for _, reg := range regions {
		amount, err := parseGermanFloat(reg.amountStr)
		if err != nil {
			continue
		}
		desc := strings.TrimSpace(strings.Join(reg.descTokens, " "))
		ref := strings.TrimSpace(strings.Join(reg.contTokens, " "))

		// In ING PDFs the first descToken is the booking type (e.g. "Lastschrift",
		// "Gutschrift", "Überweisung") and subsequent tokens are the counterparty name.
		var bankTxCode, counterpartyName string
		if len(reg.descTokens) > 0 {
			bankTxCode = strings.TrimSpace(reg.descTokens[0])
		}
		if len(reg.descTokens) > 1 {
			counterpartyName = strings.TrimSpace(strings.Join(reg.descTokens[1:], " "))
		}

		txType := entity.TransactionTypeCredit
		if amount < 0 {
			txType = entity.TransactionTypeDebit
		}
		txns = append(txns, entity.Transaction{
			BookingDate:         reg.booking,
			ValutaDate:          reg.valuta,
			Description:         desc,
			BankTransactionCode: bankTxCode,
			CounterpartyName:    counterpartyName,
			Amount:              amount,
			Currency:            "EUR",
			Type:                txType,
			Reference:           ref,
		})
	}
	return txns
}

func stripAllPageHeaders(tokens []string) []string {
	remove := make([]bool, len(tokens))
	for i, t := range tokens {
		if strings.HasPrefix(t, "ING-DiBa AG") {
			for j := i - 1; j >= 0 && !remove[j]; j-- {
				if reDate.MatchString(tokens[j]) || reAmount.MatchString(tokens[j]) {
					break
				}
				remove[j] = true
			}
			for j := i; j < len(tokens); j++ {
				remove[j] = true
				if tokens[j] == "Valuta" {
					break
				}
			}
		}
		if strings.HasPrefix(t, "Girokonto Nummer") {
			for j := i; j < len(tokens); j++ {
				remove[j] = true
				if tokens[j] == "Valuta" {
					break
				}
			}
		}
	}
	out := make([]string, 0, len(tokens))
	for i, t := range tokens {
		if !remove[i] {
			out = append(out, t)
		}
	}
	return out
}

func cutAtFooter(tokens []string) []string {
	sentinels := []string{
		"Kunden-Information",
		"Bitte beachten Sie die nachstehenden",
		"Vorliegender Freistellungsauftrag",
	}
	for i, t := range tokens {
		for _, s := range sentinels {
			if strings.Contains(t, s) {
				return tokens[:i]
			}
		}
	}
	return tokens
}

func isBoilerplate(t string) bool {
	exact := map[string]bool{
		"Buchung": true, "Valuta": true, "Betrag (EUR)": true,
		"Buchung / Verwendungszweck": true, "Datum": true, "IBAN": true,
		"BIC": true, "Seite": true, "Auszugsnummer": true, "m": true, "n": true, "/": true,
	}
	if exact[t] {
		return true
	}
	if strings.HasSuffix(t, "Euro") {
		return true
	}
	prefixes := []string{
		"ING-DiBa AG", "Girokonto Nummer", "Kontoauszug ", "Theodor-Heuss",
		"Steuernummer", "USt-IdNr", "Mitglied im", "34GKKA", "Eingerumte",
		"Eingeräumte", "Alter Saldo", "Am Sportplatz", "und Mathias", "DE23 5001", "DE23500",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

func parseGermanFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func parseGermanDate(s string) (time.Time, error) {
	return time.Parse("02.01.2006", strings.TrimSpace(s))
}
