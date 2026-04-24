package service

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCalculateSimilarity(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected float64
	}{
		{"netflix", "netflix", 1.0},
		{"netflix", "netflix.com", 0.6363636363636364}, // 1 - 4/11
		{"amazon prime", "amazon prime video", 0.6666666666666666}, // 1 - 6/18
		{"apple", "aple", 0.8}, // 1 - 1/5
		{"spotify", "spotyfy", 0.8571428571428571}, // 1 - 1/7
		{"", "test", 0.0},
		{"test", "", 0.0},
		{"abc", "def", 0.0},
		// Added cases for reconciliation
		{"paypal", "paypal.com", 0.6}, // 1 - 4/10
		{"rewe", "rewe markt", 0.4}, // 1 - 6/10
	}

	for _, tt := range tests {
		t.Run(tt.s1+" vs "+tt.s2, func(t *testing.T) {
			sim := calculateSimilarity(tt.s1, tt.s2)
			assert.InDelta(t, tt.expected, sim, 0.0001)
		})
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
	}

	for _, tt := range tests {
		t.Run(tt.s1+" vs "+tt.s2, func(t *testing.T) {
			dist := levenshtein(tt.s1, tt.s2)
			assert.Equal(t, tt.expected, dist)
		})
	}
}

func TestIsAmountWithin(t *testing.T) {
	assert.True(t, isAmountWithin(100.0, 100.005, 0.01))
	assert.False(t, isAmountWithin(100.0, 100.02, 0.01))
}

func TestIsAmountClose(t *testing.T) {
	assert.True(t, isAmountClose(100.0, 105.0, 0.051))
	assert.False(t, isAmountClose(100.0, 110.0, 0.05))
	assert.True(t, isAmountClose(0.005, 0.0, 0.1)) // Near zero case
}
