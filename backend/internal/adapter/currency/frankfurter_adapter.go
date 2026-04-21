package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cogni-cash/internal/domain/port"
)

type FrankfurterExchangeRateAdapter struct {
	baseURL string
	client  *http.Client
}

func NewFrankfurterExchangeRateAdapter() *FrankfurterExchangeRateAdapter {
	return &FrankfurterExchangeRateAdapter{
		baseURL: "https://www.frankfurter.app",
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

var _ port.CurrencyExchangeRatePort = (*FrankfurterExchangeRateAdapter)(nil)

type frankfurterResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

func (a *FrankfurterExchangeRateAdapter) GetRate(ctx context.Context, from, to string, date time.Time) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	dateStr := date.Format("2006-01-02")
	url := fmt.Sprintf("%s/%s?from=%s&to=%s", a.baseURL, dateStr, from, to)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("frankfurter: create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("frankfurter: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("frankfurter: unexpected status: %d", resp.StatusCode)
	}

	var res frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, fmt.Errorf("frankfurter: decode response: %w", err)
	}

	rate, ok := res.Rates[to]
	if !ok {
		return 0, fmt.Errorf("frankfurter: rate not found for %s", to)
	}

	return rate, nil
}
