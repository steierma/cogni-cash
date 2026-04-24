package service

import (
	"math"
	"regexp"
	"strings"
)

var (
	yearRegex      = regexp.MustCompile(`\b(20\d{2})\b`)
	monthYearRegex = regexp.MustCompile(`\b(\d{2}[/.]\d{2,4})\b`)
	// Leading reference/mandate numbers (pure digit sequences of 6+ chars)
	leadingNumericRef = regexp.MustCompile(`^\d{6,}\s+`)
	// Inline reference/mandate numbers mixed into descriptions (e.g. "00000000 IM NAMEN...")
	inlineNumericTokens = regexp.MustCompile(`\b\d{6,}\b`)
	// Common generic prefixes in bank statements
	genericPrefixes = []string{"lastschrift", "dauerauftrag", "gutschrift", "ueberweisung", "abbuchung", "zahlung"}
	// Common legal entity suffixes to strip
	legalSuffixes = []string{"gmbh", "ag", "se", "e.v.", "gdbr", "ltd", "inc", "co. kg", "ohg", "kgaa"}
)

func normalizeDescription(desc string) string {
	d := strings.ToLower(strings.TrimSpace(desc))

	// Strip leading numeric reference numbers (e.g. "0153953151 IM NAMEN..." → "IM NAMEN...")
	d = leadingNumericRef.ReplaceAllString(d, "")

	// Remove all standalone numeric tokens of 6+ digits (mandate refs, transaction IDs, etc.)
	d = inlineNumericTokens.ReplaceAllString(d, "")

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

func calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	dist := levenshtein(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	return 1.0 - float64(dist)/float64(maxLen)
}

func levenshtein(s1, s2 string) int {
	r1 := []rune(s1)
	r2 := []rune(s2)
	n := len(r1)
	m := len(r2)

	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	column := make([]int, n+1)
	for i := 1; i <= n; i++ {
		column[i] = i
	}

	for j := 1; j <= m; j++ {
		column[0] = j
		lastDiagonal := j - 1
		for i := 1; i <= n; i++ {
			oldColumnI := column[i]
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			column[i] = min3(column[i]+1, column[i-1]+1, lastDiagonal+cost)
			lastDiagonal = oldColumnI
		}
	}

	return column[n]
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= a && b <= c {
		return b
	}
	return c
}

func isAmountWithin(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func isAmountClose(a, b, tolerancePercent float64) bool {
	if b == 0 {
		return math.Abs(a) < 0.01
	}
	diff := math.Abs(a - b)
	return diff <= math.Abs(b)*tolerancePercent
}
