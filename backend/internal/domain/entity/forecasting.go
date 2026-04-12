package entity

import (
	"time"

	"github.com/google/uuid"
)

// ExcludedForecast represents a prediction that the user has explicitly chosen to hide.
type ExcludedForecast struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	ForecastID uuid.UUID `json:"forecast_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// PatternExclusion represents a rule to ignore a specific recurring transaction description/counterparty.
type PatternExclusion struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	MatchTerm string    `json:"match_term"`
	CreatedAt time.Time `json:"created_at"`
}

// ForecastPoint represents a single data point in the cash flow forecast.
type ForecastPoint struct {
	Date            time.Time          `json:"date"`
	ExpectedBalance float64            `json:"expected_balance"`
	Income          float64            `json:"income"`
	Expense         float64            `json:"expense"`
	CategoryAmounts map[string]float64 `json:"category_amounts"`
}

// PredictedTransaction represents a single future transaction that the system predicts will happen.
type PredictedTransaction struct {
	Transaction
	Probability float64 `json:"probability"`
}

// CashFlowForecast is the top-level DTO returned by the ForecastingUseCase.
type CashFlowForecast struct {
	CurrentBalance float64                `json:"current_balance"`
	TimeSeries     []ForecastPoint        `json:"time_series"`
	Predictions    []PredictedTransaction `json:"predictions"`
}
