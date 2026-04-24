package entity

import "time"

// PredictedTransaction wraps a Transaction with a probability score.
type PredictedTransaction struct {
	Transaction
	Probability float64 `json:"probability"`
}

// ForecastPoint represents a single point in the balance time series.
type ForecastPoint struct {
	Date            time.Time          `json:"date"`
	ExpectedBalance float64            `json:"expected_balance"`
	Income          float64            `json:"income"`
	Expense         float64            `json:"expense"`
	CategoryAmounts map[string]float64 `json:"category_amounts"`
}

// CashFlowForecast is the top-level forecast response.
type CashFlowForecast struct {
	CurrentBalance float64                `json:"current_balance"`
	TimeSeries     []ForecastPoint        `json:"time_series"`
	Predictions    []PredictedTransaction `json:"predictions"`
}

