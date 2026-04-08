package enablebanking

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"log/slog"

	"cogni-cash/internal/domain/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	baseURL = "https://api.enablebanking.com"
)

type Adapter struct {
	appID      string
	privateKey *rsa.PrivateKey
	logger     *slog.Logger
	client     *http.Client

	mu            sync.RWMutex
	accountsCache map[string][]entity.BankAccount
}

func NewAdapter(appID string, privateKey *rsa.PrivateKey, logger *slog.Logger) *Adapter {
	return &Adapter{
		appID:         appID,
		privateKey:    privateKey,
		logger:        logger,
		client:        &http.Client{Timeout: 30 * time.Second},
		accountsCache: make(map[string][]entity.BankAccount),
	}
}

func (a *Adapter) getJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "enablebanking.com",
		"aud": "api.enablebanking.com",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"sub": a.appID,
	})
	token.Header["kid"] = a.appID

	return token.SignedString(a.privateKey)
}

func (a *Adapter) doRequest(ctx context.Context, method, path string, bodyData io.Reader) ([]byte, error) {
	token, err := a.getJWT()
	if err != nil {
		return nil, fmt.Errorf("enablebanking: failed to sign JWT: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, method, baseURL+path, bodyData)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("enablebanking: request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("enablebanking: error status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Implementation of port.BankProvider

func (a *Adapter) GetInstitutions(ctx context.Context, userID uuid.UUID, countryCode string, isSandbox bool) ([]entity.BankInstitution, error) {
	path := fmt.Sprintf("/aspsps?country=%s", countryCode)

	if isSandbox {
		path += "&test=true" // Zwingend nötig, damit "Sample" von Enable Banking zurückgegeben wird
	}

	resp, err := a.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var res struct {
		ASPSPs []struct {
			Name    string `json:"name"`
			Title   string `json:"title"`
			Country string `json:"country"`
			Bic     string `json:"bic"`
			Logo    string `json:"logo"`
		} `json:"aspsps"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	institutions := make([]entity.BankInstitution, len(res.ASPSPs))
	for i, r := range res.ASPSPs {
		institutions[i] = entity.BankInstitution{
			ID:      r.Name,
			Name:    r.Title,
			Bic:     r.Bic,
			Logo:    r.Logo,
			Country: r.Country,
		}
	}
	return institutions, nil
}

func (a *Adapter) CreateRequisition(ctx context.Context, userID uuid.UUID, institutionID, institutionName, country, redirectURL, referenceID string, isSandbox bool) (*entity.BankConnection, error) {
	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"aspsp": map[string]string{
			"name":    institutionID,
			"country": country,
		},
		"redirect_url": redirectURL,
		"state":        referenceID,
		"access": map[string]interface{}{
			"valid_until": time.Now().Add(90 * 24 * time.Hour).Format(time.RFC3339),
			"psu_type":    "personal",
			"scopes":      []string{"accounts", "balances", "transactions"},
		},
	})

	resp, err := a.doRequest(ctx, "POST", "/auth", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	var res struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(90 * 24 * time.Hour)

	return &entity.BankConnection{
		InstitutionID:   institutionID,
		InstitutionName: institutionName,
		RequisitionID:   referenceID,
		ReferenceID:     referenceID,
		AuthLink:        res.URL,
		Status:          entity.StatusInitialized,
		ExpiresAt:       &expiresAt,
	}, nil
}

func (a *Adapter) ExchangeCodeForSession(ctx context.Context, userID uuid.UUID, code string) (string, error) {
	payload, _ := json.Marshal(map[string]string{
		"code": code,
	})

	resp, err := a.doRequest(ctx, "POST", "/sessions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	var res struct {
		SessionID string `json:"session_id"`
		Accounts  []struct {
			UID       string `json:"uid"`
			AccountID struct {
				IBAN string `json:"iban"`
				BBAN string `json:"bban"`
			} `json:"account_id"`
			Currency        string `json:"currency"`
			Name            string `json:"name"`
			CashAccountType string `json:"cash_account_type"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return "", err
	}

	// Convert to domain entities
	bankAccounts := make([]entity.BankAccount, len(res.Accounts))
	for i, r := range res.Accounts {
		// Fallback to BBAN if IBAN is not provided
		iban := r.AccountID.IBAN
		if iban == "" {
			iban = r.AccountID.BBAN
		}

		// Detect account type based on PSD2 CashAccountType
		// CACC=Current, CARD=Card, SVGS=Savings, MONE=Money Market
		accType := entity.StatementTypeGiro
		switch r.CashAccountType {
		case "CARD":
			accType = entity.StatementTypeCreditCard
		case "SVGS", "MONE":
			accType = entity.StatementTypeExtraAccount
		}

		bankAccounts[i] = entity.BankAccount{
			ProviderAccountID: r.UID, // The UID is the actual string identifier for the account
			IBAN:              iban,
			Name:              r.Name,
			Currency:          r.Currency,
			AccountType:       accType,
		}
	}

	// Cache the accounts for the upcoming FetchAccounts call
	a.mu.Lock()
	a.accountsCache[res.SessionID] = bankAccounts
	a.mu.Unlock()

	return res.SessionID, nil
}

func (a *Adapter) GetRequisitionStatus(ctx context.Context, userID uuid.UUID, requisitionID string) (entity.ConnectionStatus, error) {
	resp, err := a.doRequest(ctx, "GET", "/sessions/"+requisitionID, nil)
	if err != nil {
		return entity.StatusFailed, err
	}

	var res struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return entity.StatusFailed, err
	}

	switch res.Status {
	case "RE": // Received / Active
		return entity.StatusLinked, nil
	case "ER":
		return entity.StatusFailed, nil
	default:
		return entity.StatusInitialized, nil
	}
}

func (a *Adapter) FetchAccounts(ctx context.Context, userID uuid.UUID, requisitionID string) ([]entity.BankAccount, error) {
	a.mu.Lock()
	accounts, exists := a.accountsCache[requisitionID]
	if exists {
		// Clean up the cache after fetching
		delete(a.accountsCache, requisitionID)
	}
	a.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("enablebanking: no accounts cached for session id: %s", requisitionID)
	}

	return accounts, nil
}

func extractRemittanceData(raw string) (creditorID, mandateReference, description, location string) {
        description = raw // Default fallback

        if strings.Contains(raw, "remittanceinformation:") {
                // Extract Creditor ID safely
                cidStart := strings.Index(raw, "creditorid:")
                if cidStart != -1 {
                        cidStart += len("creditorid:")
                        cidEnd := strings.Index(raw[cidStart:], ",")
                        if cidEnd != -1 {
                                creditorID = raw[cidStart : cidStart+cidEnd]
                        } else {
                                creditorID = raw[cidStart:]
                        }
                }

                // Extract Mandate Reference
                mrStart := strings.Index(raw, "mandatereference:")
                if mrStart != -1 {
                        mrStart += len("mandatereference:")
                        mrEnd := strings.Index(raw[mrStart:], ",")
                        if mrEnd != -1 {
                                mandateReference = raw[mrStart : mrStart+mrEnd]
                        } else {
                                mandateReference = raw[mrStart:]
                        }
                }

                // Extract the actual description text
                remStart := strings.Index(raw, "remittanceinformation:")
                if remStart != -1 {
                        description = raw[remStart+len("remittanceinformation:"):]

                        // Regex to find city names before "Datum" or " DE "
                        // Matches letters (including German umlauts), allowing spaces/hyphens inside,
                        // right before the keyword "Datum" or isolated "DE".
                        locRegex := regexp.MustCompile(`(?i)([A-ZÄÖÜa-zäöüß]+(?:[\s\-][A-ZÄÖÜa-zäöüß]+)*)\s*(?:Datum|\s+DE\s+)`)
                        matches := locRegex.FindStringSubmatch(description)
                        if len(matches) > 1 {
                                location = strings.TrimSpace(matches[1])
                        }
                }
        }

        return strings.TrimSpace(creditorID), strings.TrimSpace(mandateReference), strings.TrimSpace(description), location
}
func (a *Adapter) FetchTransactions(ctx context.Context, userID uuid.UUID, providerAccountID string, dateFrom *time.Time, dateTo *time.Time) ([]entity.Transaction, float64, error) {
	if providerAccountID == "dummy_acc_id" || strings.HasPrefix(providerAccountID, "mock_") {
		return nil, 0, fmt.Errorf("enablebanking: invalid account id '%s' (this looks like mock data)", providerAccountID)
	}

	a.logger.Debug("fetching transactions from Enable Banking", "account_id", providerAccountID)

	urlPath := "/accounts/" + providerAccountID + "/transactions"
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
		a.logger.Error("failed to fetch transactions from Enable Banking", "account_id", providerAccountID, "error", err)
		return nil, 0, err
	}
	a.logger.Debug("successfully fetched transactions from Enable Banking", "account_id", providerAccountID, "response", resp)

	var res struct {
		Transactions []struct {
			TransactionID  string `json:"transaction_id"`
			EntryReference string `json:"entry_reference"`
			BookingDate    string `json:"booking_date"`
			ValueDate      string `json:"value_date"`
			Status         string `json:"status"` // "BOOK" (Booked) or "PDNG" (Pending)
			AmountObj      struct {
				Amount   string `json:"amount"`
				Currency string `json:"currency"`
			} `json:"transaction_amount"`
			CreditDebitIndicator string   `json:"credit_debit_indicator"` // "CRDT" or "DBIT"
			Remittance           string   `json:"remittance_information_unstructured"`
			RemittanceArray      []string `json:"remittance_information"`
			CreditorName         string   `json:"creditor_name"`
			DebtorName           string   `json:"debtor_name"`
			Creditor             struct {
				Name string `json:"name"`
			} `json:"creditor"`
			Debtor struct {
				Name string `json:"name"`
			} `json:"debtor"`
		} `json:"transactions"`
	}
	if err := json.Unmarshal(resp, &res); err != nil {
		return nil, 0, err
	}

	// Fetch balance
	balResp, err := a.doRequest(ctx, "GET", "/accounts/"+providerAccountID+"/balances", nil)
	var balance float64
	if err == nil {
		var balRes struct {
			Balances []struct {
				BalanceAmount struct {
					Amount   string `json:"amount"`
					Currency string `json:"currency"`
				} `json:"balance_amount"`
			} `json:"balances"`
		}
		if err := json.Unmarshal(balResp, &balRes); err == nil && len(balRes.Balances) > 0 {
			balance, _ = strconv.ParseFloat(balRes.Balances[0].BalanceAmount.Amount, 64)
		}
	} else {
		a.logger.Warn("failed to fetch balance from Enable Banking (skipping)", "account_id", providerAccountID, "error", err)
	}

	var transactions []entity.Transaction
	for _, r := range res.Transactions {
		// 1. FILTER: Only process booked/real transactions
		if r.Status == "PDNG" {
			continue
		}

		// 2. AMOUNT & SIGN
		amt, _ := strconv.ParseFloat(r.AmountObj.Amount, 64)
		txnType := "credit"

		if r.CreditDebitIndicator == "DBIT" {
			txnType = "debit"
			amt = -amt
		} else if r.CreditDebitIndicator == "" && amt < 0 {
			txnType = "debit"
		}

		bookingDate, _ := time.Parse("2006-01-02", r.BookingDate)
		valutaDate, _ := time.Parse("2006-01-02", r.ValueDate)
		if valutaDate.IsZero() {
			valutaDate = bookingDate
		}

		// 3. DESCRIPTION, LOCATION & COUNTERPARTY
		rawRemittance := r.Remittance
		if rawRemittance == "" && len(r.RemittanceArray) > 0 {
			rawRemittance = strings.Join(r.RemittanceArray, " ")
		}

		creditorID, mandateRef, parsedDesc, location := extractRemittanceData(rawRemittance)

		cName := r.CreditorName
		if cName == "" {
			cName = r.Creditor.Name
		}
		dName := r.DebtorName
		if dName == "" {
			dName = r.Debtor.Name
		}

		desc := parsedDesc
		if desc == "" {
			if txnType == "debit" {
				desc = cName
			} else {
				desc = dName
			}
		}

		// Counterparty: creditor for debits, debtor for credits
		counterparty := cName
		if txnType == "credit" {
			counterparty = dName
		}

		// 4. REFERENCE
		ref := creditorID
		if ref == "" {
			ref = r.TransactionID
			if ref == "" {
				ref = r.EntryReference
			}
		}

		transactions = append(transactions, entity.Transaction{
			BookingDate:        bookingDate,
			ValutaDate:         valutaDate,
			Description:        desc,
			Location:           location,
			Amount:             amt,
			Currency:           r.AmountObj.Currency,
			Type:               entity.TransactionType(txnType),
			Reference:          ref,
			MandateReference:   mandateRef,
			CounterpartyName:   counterparty,
		})
	}

	return transactions, balance, nil
}
