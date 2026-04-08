package cariad

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cogni-cash/internal/domain/entity"

	"log/slog"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

var (
	rePeriod   = regexp.MustCompile(`Monatsabrechnung für ([A-Za-zäöüÄÖÜß]+) (\d{4})`)
	reTaxClass = regexp.MustCompile(`Steuerklasse\s*/\s*Faktor\s*(\d)`)

	// reAmountGlobal finds German-formatted amounts anywhere in a text block, handling trailing minuses
	reAmountGlobal = regexp.MustCompile(`-?\d{1,3}(?:\.\d{3})*,\d{2}-?`)
)

// Parser implements a port for CARIAD/VW PDF payslips.
type Parser struct {
	Logger *slog.Logger
}

func NewParser(logger *slog.Logger) *Parser {
	return &Parser{Logger: logger}
}

func (p *Parser) Parse(_ context.Context, _ uuid.UUID, fileBytes []byte) (entity.Payslip, error) {
	p.Logger.Info("Parsing CARIAD Payslip PDF")
	raw, err := extractRawText(fileBytes)
	if err != nil {
		p.Logger.Error("Failed to extract text from Payslip", "error", err)
		return entity.Payslip{}, fmt.Errorf("cariad parser: extract text: %w", err)
	}

	payslip := entity.Payslip{
		EmployerName: "CARIAD SE", // Default employer for this parser
	}

	// 1. Extract Header Metadata
	if m := rePeriod.FindStringSubmatch(raw); len(m) == 3 {
		payslip.PeriodMonthNum = parseMonthNameToNum(m[1])
		payslip.PeriodYear, _ = strconv.Atoi(m[2])
	}
	if m := reTaxClass.FindStringSubmatch(raw); len(m) == 2 {
		payslip.TaxClass = m[1]
	}

	// 2. Extract Financial Data using Window Scanning
	// Map Gesamtbrutto -> GrossPay
	payslip.GrossPay = extractHighestAmountAfter(raw, "GESAMTBRUTTOENTGELT", 300)

	// Nettoentgelt can appear twice (once as subtotal, once as final with EBV). Highest is the final net.
	// Map Nettoentgelt -> NetPay
	payslip.NetPay = extractHighestAmountAfter(raw, "Nettoentgelt", 300)

	// Auszahlungsbetrag is the very last positive number in the accounting block before the Freizeit table.
	// Map Auszahlungsbetrag -> PayoutAmount
	payslip.PayoutAmount = extractAuszahlungsbetrag(raw)

	// 3. Extract Leasing / Mietrate (Find the first negative amount after the keyword)
	// Map MietrateLeasing -> CustomDeductions
	payslip.CustomDeductions = extractFirstNegativeAmountAfter(raw, "Mietrate/ Leasing", 300)

	// 4. Extract Bonuses (formerly Sonderzahlungen)
	payslip.Bonuses = extractSonderzahlungen(raw)

	p.Logger.Info("Successfully parsed Payslip", "monthNum", payslip.PeriodMonthNum, "net_pay", payslip.NetPay)
	return payslip, nil
}

// ---- text extraction -------------------------------------------------------

func extractRawText(fileBytes []byte) (string, error) {
	readerAt := bytes.NewReader(fileBytes)
	r, err := pdf.NewReader(readerAt, int64(len(fileBytes)))
	if err != nil {
		return "", err
	}

	var rawBuilder strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		rawBuilder.WriteString(text)
		rawBuilder.WriteString("\n")
	}
	return rawBuilder.String(), nil
}

// ---- targeted window scanners ----------------------------------------------

func extractHighestAmountAfter(raw, keyword string, window int) float64 {
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(keyword))
	matches := re.FindAllStringIndex(raw, -1)

	var maxVal float64
	for _, match := range matches {
		start := match[1]
		end := start + window
		if end > len(raw) {
			end = len(raw)
		}
		slice := raw[start:end]

		for _, amtStr := range reAmountGlobal.FindAllString(slice, -1) {
			val, _ := parseGermanFloat(amtStr)
			if val > maxVal {
				maxVal = val
			}
		}
	}
	return maxVal
}

func extractAuszahlungsbetrag(raw string) float64 {
	idx := strings.Index(raw, "Auszahlungsbetrag")
	if idx == -1 {
		return 0
	}

	// Create a slice from "Auszahlungsbetrag" up to the start of the Freizeit section
	endMarker := "Freizeit in"
	endIdx := strings.Index(raw[idx:], endMarker)

	var slice string
	if endIdx != -1 {
		slice = raw[idx : idx+endIdx]
	} else {
		slice = raw[idx:]
		if len(slice) > 800 {
			slice = slice[:800] // fallback boundary
		}
	}

	var lastPos float64
	for _, amtStr := range reAmountGlobal.FindAllString(slice, -1) {
		val, _ := parseGermanFloat(amtStr)
		if val > 0 {
			lastPos = val
		}
	}
	return lastPos
}

func extractFirstNegativeAmountAfter(raw, keyword string, window int) float64 {
	idx := strings.Index(raw, keyword)
	if idx == -1 {
		return 0
	}

	end := idx + window
	if end > len(raw) {
		end = len(raw)
	}
	slice := raw[idx:end]

	for _, amtStr := range reAmountGlobal.FindAllString(slice, -1) {
		val, _ := parseGermanFloat(amtStr)
		if val < 0 {
			return val
		}
	}
	return 0
}

func extractSonderzahlungen(raw string) []entity.Bonus {
	var sz []entity.Bonus
	keywords := []string{
		"Einmalzahlung",
		"13. Monatsentgelt",
		"Tariferg.",
		"TZUG Zusatzbetrag",
		"TZUG-Tarifl. Zusatzgeld",
		"Urlaubsgeld",
	}
	seen := make(map[float64]bool)

	for _, kw := range keywords {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(kw))
		matches := re.FindAllStringIndex(raw, -1)

		for _, match := range matches {
			// Bonuses amounts can sometimes be extracted slightly *before* the keyword
			start := match[0] - 150
			if start < 0 {
				start = 0
			}
			end := match[1] + 150
			if end > len(raw) {
				end = len(raw)
			}

			slice := raw[start:end]
			for _, amtStr := range reAmountGlobal.FindAllString(slice, -1) {
				val, _ := parseGermanFloat(amtStr)

				// Filter out minor amounts (hours, days, small allowances)
				if val > 100.0 && !seen[val] {
					desc := kw
					if kw == "13. Monatsentgelt" || kw == "Monatsentgelt" {
						desc = "13. Monatsentgelt"
					}
					sz = append(sz, entity.Bonus{
						Description: desc,
						Amount:      val,
					})
					seen[val] = true
				}
			}
		}
	}
	return sz
}

// ---- helpers ---------------------------------------------------------------

func parseGermanFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	isNegative := strings.HasSuffix(s, "-") || strings.HasPrefix(s, "-")

	s = strings.TrimSuffix(s, "-")
	s = strings.TrimPrefix(s, "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if isNegative {
		val = -val
	}
	return val, nil
}

func parseMonthNameToNum(monthStr string) int {
	monthStr = strings.ToLower(strings.TrimSpace(monthStr))
	switch monthStr {
	case "januar", "january", "jan":
		return 1
	case "februar", "february", "feb":
		return 2
	case "märz", "maerz", "march", "mar":
		return 3
	case "april", "apr":
		return 4
	case "mai", "may":
		return 5
	case "juni", "june", "jun":
		return 6
	case "juli", "july", "jul":
		return 7
	case "august", "aug":
		return 8
	case "september", "sep":
		return 9
	case "oktober", "october", "okt", "oct":
		return 10
	case "november", "nov":
		return 11
	case "dezember", "december", "dez", "dec":
		return 12
	default:
		return 0 // Safely return 0 if unrecognized
	}
}
