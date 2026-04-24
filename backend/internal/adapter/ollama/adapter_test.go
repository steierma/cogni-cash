package llm_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	llm "cogni-cash/internal/adapter/ollama"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type mockSettingsRepo struct {
	settings map[string]string
}

func (m *mockSettingsRepo) Get(_ context.Context, key string, _ uuid.UUID) (string, error) {
	return m.settings[key], nil
}

func (m *mockSettingsRepo) Set(_ context.Context, key, value string, _ uuid.UUID, isSensitive bool) error {
	m.settings[key] = value
	return nil
}

func (m *mockSettingsRepo) GetAll(_ context.Context, _ uuid.UUID) (map[string]string, error) {
	return m.settings, nil
}

// TestAdapter_CategorizeImage_Gemini verifies that CategorizeInvoiceImage sends the
// image as an inline_data part to the Gemini endpoint and parses the response.
func TestAdapter_CategorizeImage_Gemini(t *testing.T) {
	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "key=test-token") {
			t.Errorf("expected api key in query, got %s", r.URL.RawQuery)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)

		contents := payload["contents"].([]interface{})
		parts := contents[0].(map[string]interface{})["parts"].([]interface{})

		hasImage := false
		for _, p := range parts {
			part := p.(map[string]interface{})
			if inlineRaw, ok := part["inline_data"]; ok {
				hasImage = true
				inlineData := inlineRaw.(map[string]interface{})
				if inlineData["mime_type"] != "image/png" {
					t.Errorf("expected image/png, got %v", inlineData["mime_type"])
				}
				if inlineData["data"] == "" {
					t.Error("expected base64 image data")
				}
			}
		}
		if !hasImage {
			t.Error("expected inline_data image part in Gemini request")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"category_name\":\"Travel\",\"vendor_name\":\"Uber\",\"amount\":15.0,\"currency\":\"EUR\",\"invoice_date\":\"2024-01-01\",\"description\":\"Ride\"}"}]}}]}`))
	}))
	defer server.Close()

	settings := &mockSettingsRepo{
		settings: map[string]string{
			"llm_api_url":   server.URL + "/googleapis.com",
			"llm_api_token": "test-token",
			"llm_model":     "gemini-1.5-flash",
		},
	}

	adapter := llm.NewAdapter(settings, nil, slog.Default())

	res, err := adapter.CategorizeInvoiceImage(context.Background(), userID, "test_file.png", "image/png", []byte("fake-image"), []string{"Travel", "Food"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.InvoiceName != "Travel" {
		t.Errorf("expected Travel, got %s", res.InvoiceName)
	}
	if res.VendorName != "Uber" {
		t.Errorf("expected Uber, got %s", res.VendorName)
	}
}

// TestAdapter_CategorizeImage_OllamaFallback verifies that when the LLM URL is
// an Ollama instance (not googleapis.com) CategorizeInvoiceImage falls back to a
// text-only prompt and still returns a valid result.
func TestAdapter_CategorizeImage_OllamaFallback(t *testing.T) {
	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected /api/generate, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"{\"category_name\":\"Food\",\"vendor_name\":\"McDonalds\",\"amount\":10.0,\"currency\":\"EUR\",\"invoice_date\":\"2024-01-02\",\"description\":\"Burger\"}"}`))
	}))
	defer server.Close()

	settings := &mockSettingsRepo{
		settings: map[string]string{
			"llm_api_url":   server.URL,
			"llm_api_token": "test-token",
			"llm_model":     "llava",
		},
	}

	adapter := llm.NewAdapter(settings, nil, slog.Default())

	res, err := adapter.CategorizeInvoiceImage(context.Background(), userID, "test_file.png", "image/png", []byte("fake-image"), []string{"Travel", "Food"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.InvoiceName != "Food" {
		t.Errorf("expected Food, got %s", res.InvoiceName)
	}
}

// TestAdapter_ParsePayslip_Gemini verifies that ParsePayslip
// sends the document as an inline_data part and unmarshals the payslip response.
func TestAdapter_ParsePayslip_Gemini(t *testing.T) {
	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)

		contents := payload["contents"].([]interface{})
		parts := contents[0].(map[string]interface{})["parts"].([]interface{})

		hasImage := false
		for _, p := range parts {
			part := p.(map[string]interface{})
			if _, ok := part["inline_data"]; ok {
				hasImage = true
			}
		}
		if !hasImage {
			t.Error("expected inline_data part in Gemini payslip request")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"period_month_num\":5,\"period_year\":2024,\"gross_pay\":5000.0,\"net_pay\":3500.0}"}]}}]}`))
	}))
	defer server.Close()

	settings := &mockSettingsRepo{
		settings: map[string]string{
			"llm_api_url":   server.URL + "/googleapis.com",
			"llm_api_token": "test-token",
		},
	}

	adapter := llm.NewAdapter(settings, nil, slog.Default())

	payslip, err := adapter.ParsePayslip(context.Background(), userID, "payslip.png", "image/png", []byte("fake-image"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payslip.GrossPay != 5000.0 {
		t.Errorf("expected gross_pay 5000, got %f", payslip.GrossPay)
	}
}

func TestAdapter_CurrencyHandling(t *testing.T) {
	userID := uuid.New()

	t.Run("Decouples currency in prompt with fallback", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			_ = json.Unmarshal(body, &payload)
			
			prompt := ""
			if p, ok := payload["prompt"].(string); ok {
				prompt = p
			} else {
				contents := payload["contents"].([]interface{})
				parts := contents[0].(map[string]interface{})["parts"].([]interface{})
				prompt = parts[0].(map[string]interface{})["text"].(string)
			}

			if !strings.Contains(prompt, "USD") {
				t.Errorf("expected prompt to contain USD hint, got: %s", prompt)
			}

			w.Header().Set("Content-Type", "application/json")
			// Return an invalid currency to test normalization/fallback
			_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"category_name\":\"Travel\",\"vendor_name\":\"Uber\",\"amount\":15.0,\"currency\":\"US Dollars\",\"invoice_date\":\"2024-01-01\",\"description\":\"Ride\"}"}]}}]}`))
		}))
		defer server.Close()

		settings := &mockSettingsRepo{
			settings: map[string]string{
				"llm_api_url":   server.URL + "/googleapis.com",
				"llm_api_token": "test-token",
				"currency":      "USD",
			},
		}

		adapter := llm.NewAdapter(settings, nil, slog.Default())
		res, err := adapter.CategorizeInvoice(context.Background(), userID, port.CategorizationRequest{
			RawText: "Spent some money",
			Categories: []string{"Travel"},
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Currency != "USD" {
			t.Errorf("expected fallback to USD for invalid 'US Dollars', got %s", res.Currency)
		}
	})

	t.Run("Normalizes valid currency case", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"category_name\":\"Food\",\"vendor_name\":\"McD\",\"amount\":5.0,\"currency\":\"gbp\",\"invoice_date\":\"2024-01-01\",\"description\":\"Burger\"}"}]}}]}`))
		}))
		defer server.Close()

		settings := &mockSettingsRepo{
			settings: map[string]string{
				"llm_api_url":   server.URL + "/googleapis.com",
				"llm_api_token": "test-token",
			},
		}

		adapter := llm.NewAdapter(settings, nil, slog.Default())
		res, err := adapter.CategorizeInvoice(context.Background(), userID, port.CategorizationRequest{
			RawText: "Burger",
			Categories: []string{"Food"},
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Currency != "GBP" {
			t.Errorf("expected GBP, got %s", res.Currency)
		}
	})

	t.Run("Bank Statement inherits currency", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"account_holder\":\"John Doe\",\"iban\":\"DE123\",\"currency\":\"CHF\",\"statement_date\":\"2024-01-01\",\"transactions\":[{\"booking_date\":\"2024-01-01\",\"amount\":-10.0,\"description\":\"Test\"}]}"}]}}]}`))
		}))
		defer server.Close()

		settings := &mockSettingsRepo{
			settings: map[string]string{
				"llm_api_url":   server.URL + "/googleapis.com",
				"llm_api_token": "test-token",
			},
		}

		adapter := llm.NewAdapter(settings, nil, slog.Default())
		stmt, err := adapter.ParseBankStatement(context.Background(), userID, "stmt.pdf", "application/pdf", []byte("fake-pdf"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stmt.Currency != "CHF" {
			t.Errorf("expected CHF, got %s", stmt.Currency)
		}
		if len(stmt.Transactions) == 0 {
			t.Fatal("expected 1 transaction")
		}
		if stmt.Transactions[0].Currency != "CHF" {
			t.Errorf("expected transaction to inherit CHF, got %s", stmt.Transactions[0].Currency)
		}
	})
}
