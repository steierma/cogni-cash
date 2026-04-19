package service

import (
	"regexp"
	"strings"
)

var (
	yearRegex      = regexp.MustCompile(`\b(20\d{2})\b`)
	monthYearRegex = regexp.MustCompile(`\b(\d{2}[/\.]\d{2,4})\b`)
	// Common generic prefixes in bank statements
	genericPrefixes = []string{"lastschrift", "dauerauftrag", "gutschrift", "ueberweisung", "abbuchung", "zahlung"}
	// Common legal entity suffixes to strip
	legalSuffixes = []string{"gmbh", "ag", "se", "e.v.", "gdbr", "ltd", "inc", "co. kg", "ohg", "kgaa"}
)

func normalizeDescription(desc string) string {
	d := strings.ToLower(strings.TrimSpace(desc))

	// Remove generic prefixes recursively
	changed := true
	for changed {
		changed = false
		for _, p := range genericPrefixes {
			if strings.HasPrefix(d, p) {
				d = strings.TrimSpace(strings.TrimPrefix(d, p))
				changed = true
			}
		}
	}

	// Remove legal entity suffixes
	for _, s := range legalSuffixes {
		if strings.HasSuffix(d, s) {
			d = strings.TrimSpace(strings.TrimSuffix(d, s))
			break // Only remove one suffix
		}
	}

	// Remove year patterns like "2024" or "2023"
	d = yearRegex.ReplaceAllString(d, "")
	// Remove month/year patterns like "05/24"
	d = monthYearRegex.ReplaceAllString(d, "")
	// Clean up extra spaces
	d = strings.Join(strings.Fields(d), " ")

	// Remove leading/trailing special characters that might be left over from prefix removal
	d = strings.Trim(d, " -/*:.,")

	if len(d) > 60 {
		d = d[:60]
	}
	return strings.TrimSpace(d)
}
func getIntervalFromDays(days float64, tolerance float64) int {
	// Monthly: ~30 days
	if days >= 30-tolerance && days <= 30+tolerance {
		return 1
	}
	// Quarterly: ~91 days
	if days >= 91-tolerance && days <= 91+tolerance {
		return 3
	}
	// Half-yearly: ~182 days
	if days >= 182-tolerance && days <= 182+tolerance {
		return 6
	}
	// Yearly: ~365 days
	if days >= 365-tolerance && days <= 365+tolerance {
		return 12
	}
	return 0
}
