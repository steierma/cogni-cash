package gocardless

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"cogni-cash/internal/domain/entity"
)

const (
	baseURL = "https://bankaccountdata.gocardless.com/api/v2"
)

type Adapter struct {
	secretID  string
	secretKey string
	logger    *slog.Logger
	client    *http.Client

	// Simple token cache
	token       string
	tokenExpiry time.Time
}

func NewAdapter(secretID, secretKey string, logger *slog.Logger) *Adapter {
	return &Adapter{
		secretID:  secretID,
		secretKey: secretKey,
		logger:    logger,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Internal token management

func (a *Adapter) getAuthToken(ctx context.Context) (string, error) {
	if a.secretID == "" || a.secretKey == "" {
		return "", fmt.Errorf("gocardless: missing API credentials (GOCARDLESS_SECRET_ID/KEY)")
	}
	if a.token != "" && time.Now().Before(a.tokenExpiry) {
		return a.token, nil
	}

	payload := map[string]string{
		"secret_id":  a.secretID,
		"secret_key": a.secretKey,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/token/new/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gocardless: token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gocardless: auth failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var res struct {
		Access string `json:"access"`
		Expiry int    `json:"access_expires"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("gocardless: decode token: %w", err)
	}

	a.token = res.Access
	a.tokenExpiry = time.Now().Add(time.Duration(res.Expiry-60) * time.Second)
	return a.token, nil
}

func (a *Adapter) doRequest(ctx context.Context, method, path string, bodyData interface{}) ([]byte, error) {
	token, err := a.getAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if bodyData != nil {
		jsonBytes, _ := json.Marshal(bodyData)
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	req, _ := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gocardless: request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gocardless: error status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Implementation of port.BankProvider

func (a *Adapter) GetInstitutions(ctx context.Context, countryCode string, isSandbox bool) ([]entity.BankInstitution, error) {
	path := fmt.Sprintf("/institutions/?country=%s", countryCode)
	resp, err := a.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var res []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Bic     string `json:"bic"`
		Logo    string `json:"logo"`
		Country string `json:"countries"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	institutions := make([]entity.BankInstitution, len(res))
	for i, r := range res {
		institutions[i] = entity.BankInstitution{
			ID:      r.ID,
			Name:    r.Name,
			Bic:     r.Bic,
			Logo:    r.Logo,
			Country: r.Country,
		}
	}
	return institutions, nil
}

func (a *Adapter) CreateRequisition(ctx context.Context, institutionID, country, redirectURL, referenceID string, isSandbox bool) (*entity.BankConnection, error) {
	payload := map[string]interface{}{
		"redirect":       redirectURL,
		"institution_id": institutionID,
		"reference":      referenceID,
		"user_language":  country,
	}

	resp, err := a.doRequest(ctx, "POST", "/requisitions/", payload)
	if err != nil {
		return nil, err
	}

	var res struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Link    string `json:"link"`
		Created string `json:"created"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	createdAt, _ := time.Parse(time.RFC3339, res.Created)
	expiresAt := createdAt.Add(90 * 24 * time.Hour)

	return &entity.BankConnection{
		InstitutionID: institutionID,
		RequisitionID: res.ID,
		ReferenceID:   referenceID,
		Status:        entity.StatusInitialized,
		AuthLink:      res.Link,
		CreatedAt:     createdAt,
		ExpiresAt:     &expiresAt,
	}, nil
}

func (a *Adapter) ExchangeCodeForSession(ctx context.Context, code string) (string, error) {
	return code, nil
}

func (a *Adapter) GetRequisitionStatus(ctx context.Context, requisitionID string) (entity.ConnectionStatus, error) {
	resp, err := a.doRequest(ctx, "GET", "/requisitions/"+requisitionID+"/", nil)
	if err != nil {
		return entity.StatusFailed, err
	}

	var res struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return entity.StatusFailed, err
	}

	// Nordigen statuses: CR (Created), LN (Linked), EX (Expired), RJ (Rejected), etc.
	switch res.Status {
	case "LN":
		return entity.StatusLinked, nil
	case "EX":
		return entity.StatusExpired, nil
	case "RJ", "ER":
		return entity.StatusFailed, nil
	default:
		return entity.StatusInitialized, nil
	}
}

func (a *Adapter) FetchAccounts(ctx context.Context, requisitionID string) ([]entity.BankAccount, error) {
	resp, err := a.doRequest(ctx, "GET", "/requisitions/"+requisitionID+"/", nil)
	if err != nil {
		return nil, err
	}

	var res struct {
		Accounts []string `json:"accounts"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	var bankAccounts []entity.BankAccount
	for _, accID := range res.Accounts {
		// Fetch details for each account
		accResp, err := a.doRequest(ctx, "GET", "/accounts/"+accID+"/", nil)
		if err != nil {
			a.logger.Error("failed to fetch account details", "account_id", accID, "error", err)
			continue
		}

		var accRes struct {
			IBAN     string `json:"iban"`
			Currency string `json:"currency"`
			Owner    string `json:"owner_name"`
			Product  string `json:"product"`
		}
		if err := json.Unmarshal(accResp, &accRes); err != nil {
			a.logger.Error("failed to unmarshal account details", "account_id", accID, "error", err)
			continue
		}

		bankAccounts = append(bankAccounts, entity.BankAccount{
			ProviderAccountID: accID,
			IBAN:              accRes.IBAN,
			Name:              accRes.Product,
			Currency:          accRes.Currency,
		})
	}

	return bankAccounts, nil
}

func (a *Adapter) FetchTransactions(ctx context.Context, providerAccountID string, dateFrom *time.Time, dateTo *time.Time) ([]entity.Transaction, float64, error) {
	// 1. Fetch Transactions
	urlPath := "/accounts/" + providerAccountID + "/transactions/"
	var query []string
	if dateFrom != nil {
		query = append(query, "date_from="+dateFrom.Format("2006-01-02"))
	}
	if dateTo != nil {
		query = append(query, "date_to="+dateTo.Format("2006-01-02"))
	}
	if len(query) > 0 {
		urlPath += "?" + strings.Join(query, "&")
	}

	resp, err := a.doRequest(ctx, "GET", urlPath, nil)
	if err != nil {
		return nil, 0, err
	}

	var res struct {
		Transactions struct {
			Booked []struct {
				ID                string `json:"transactionId"`
				BookingDate       string `json:"bookingDate"`
				ValueDate         string `json:"valueDate"`
				TransactionAmount struct {
					Amount   string `json:"amount"`
					Currency string `json:"currency"`
				} `json:"transactionAmount"`
				RemittanceInformationUnstructured string `json:"remittanceInformationUnstructured"`
				CreditorName                      string `json:"creditorName"`
				DebtorName                        string `json:"debtorName"`
			} `json:"booked"`
		} `json:"transactions"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, 0, err
	}

	// 2. Fetch Balance (to update the account)
	var currentBalance float64
	balResp, err := a.doRequest(ctx, "GET", "/accounts/"+providerAccountID+"/balances/", nil)
	if err == nil {
		var balRes struct {
			Balances []struct {
				BalanceAmount struct {
					Amount   string `json:"amount"`
					Currency string `json:"currency"`
				} `json:"balanceAmount"`
				BalanceType string `json:"balanceType"`
			} `json:"balances"`
		}

		if err := json.Unmarshal(balResp, &balRes); err == nil && len(balRes.Balances) > 0 {
			// Try to find "interimAvailable" or "closingBooked"
			if parsedBal, err := strconv.ParseFloat(balRes.Balances[0].BalanceAmount.Amount, 64); err == nil {
				currentBalance = parsedBal
			}
		}
	} else {
		a.logger.Warn("failed to fetch balance", "account_id", providerAccountID, "error", err)
	}

	var transactions []entity.Transaction
	for _, t := range res.Transactions.Booked {
		amt, _ := strconv.ParseFloat(t.TransactionAmount.Amount, 64)

		bookingDate, _ := time.Parse("2006-01-02", t.BookingDate)
		valutaDate, _ := time.Parse("2006-01-02", t.ValueDate)

		desc := t.RemittanceInformationUnstructured
		if desc == "" {
			if amt < 0 {
				desc = t.CreditorName
			} else {
				desc = t.DebtorName
			}
		}

		transactions = append(transactions, entity.Transaction{
			BookingDate: bookingDate,
			ValutaDate:  valutaDate,
			Description: desc,
			Amount:      amt,
			Currency:    t.TransactionAmount.Currency,
			Type:        entity.TransactionType(fmt.Sprintf("%v", amt >= 0)), // temp logic
			Reference:   t.ID,
		})
	}

	return transactions, currentBalance, nil
}
