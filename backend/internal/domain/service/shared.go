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
	// We want to detect if the gap is roughly a multiple of 1 month (approx 30 days).
	// This handles missing months in the middle of a sequence.
	
	if days < 20 { // Too short for any recurring interval we care about
		return 0
	}

	// Helper to check if a value is within tolerance of a target
	isClose := func(val, target, tol float64) bool {
		return val >= target-tol && val <= target+tol
	}

	// 1. Monthly (or multiple thereof: Quarterly=3, Half-Yearly=6, Yearly=12)
	for m := 1; m <= 12; m++ {
		target := float64(m) * 30.44 // Average month length
		// Increase tolerance for longer gaps
		effectiveTol := tolerance + (float64(m) * 0.5) 
		
		if isClose(days, target, effectiveTol) {
			return m
		}
	}

	return 0
}
