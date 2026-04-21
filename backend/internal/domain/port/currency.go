package port

import (
	"context"
	"time"
)

// CurrencyExchangeRatePort defines the interface for fetching historical and current exchange rates.
type CurrencyExchangeRatePort interface {
	// GetRate returns the exchange rate from one currency to another for a specific date.
	GetRate(ctx context.Context, from, to string, date time.Time) (float64, error)
}
