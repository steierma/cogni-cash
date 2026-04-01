package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const defaultModel = "gemini-1.5-flash"

const defaultSinglePromptTemplate = `Categorize the following invoice text. 
Use EXACTLY ONE category from: [{{CATEGORIES}}].
Return ONLY a valid JSON object. Do not include explanations.

JSON Schema:
{"category_name": "string", "vendor_name": "string", "amount": 12.34, "currency": "EUR", "invoice_date": "YYYY-MM-DD", "description": "string"}

TEXT:
{{TEXT}}`

const defaultBatchPromptTemplate = `Categorize these transactions using ONLY: [{{CATEGORIES}}].
Return ONLY a valid JSON array of objects.
Each object MUST have "hash" and "category".

Here are some examples of past categorizations for reference:
{{EXAMPLES}}

DATA TO CATEGORIZE:
{{DATA}}`

const defaultStatementPromptTemplate = `Extract bank statement details from the following text.
Return ONLY a valid JSON object. Do not include explanations or markdown formatting outside of the JSON block.

JSON Schema:
{
  "account_holder": "string",
  "iban": "string",
  "currency": "EUR",
  "statement_date": "YYYY-MM-DD",
  "statement_no": 123,
  "new_balance": 1234.56,
  "transactions": [
    {
      "booking_date": "YYYY-MM-DD",
      "amount": -12.34,
      "description": "string",
      "reference": "string",
      "counterparty_name": "string",
      "counterparty_iban": "string",
      "bank_transaction_code": "string",
      "mandate_reference": "string"
    }
  ]
}

TEXT:
{{TEXT}}`

const defaultPayslipPromptTemplate = `Role: You are a precise financial data extraction system.
Task: Extract payroll information from the provided payslip and map it strictly to the JSON schema below.

Strict Extraction Rules:

No Hallucinations: Extract values exactly as they are represented. Do not calculate, guess, or infer numbers.

Missing Data: If a value is not explicitly found in the text, you must return null for that field.

Number Formatting: Convert localized number formats (e.g., 1.234,56 or 1,234.56) into standard float values (e.g., 1234.56) without thousands separators.

Date Mapping: Convert month names found in the text into their corresponding integer (e.g., "January" / "Januar" = 1, "May" / "Mai" = 5).

Output Constraint: Return ONLY raw, valid JSON. Do not wrap the output in markdown code blocks (do not use ` + "```" + `json). Do not include any conversational text, explanations, or formatting outside the JSON object.

JSON Schema Definition:
{
  "period_month_num": "integer (1-12)",
  "period_year": "integer (YYYY)",
  "employee_name": "string",
  "employer_name": "string",
  "tax_class": "string",
  "tax_id": "string",
  "gross_pay": "float",
  "net_pay": "float",
  "payout_amount": "float",
  "custom_deductions": "float or null",
  "bonuses": [{"description": "string", "amount": "float"}]
}

Source Text:
{{TEXT}}`

// --- Google Gemini Types ---

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

// geminiPart supports both plain text and inline binary blobs (multimodal).
type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inline_data,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"` // base64-encoded
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// --- Ollama Types (Legacy Support) ---

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
}

// --- Internal Mapping Types ---

type llmResponse struct {
	CategoryName string  `json:"category_name"`
	VendorName   string  `json:"vendor_name"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	InvoiceDate  string  `json:"invoice_date"`
	Description  string  `json:"description"`
}

type Adapter struct {
	settingsRepo port.SettingsRepository
	client       *http.Client
	logger       *slog.Logger
}

func NewAdapter(settingsRepo port.SettingsRepository, logger *slog.Logger) *Adapter {
	if logger == nil {
		logger = slog.Default()
	}
	return &Adapter{
		settingsRepo: settingsRepo,
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
		logger: logger,
	}
}

func (a *Adapter) getLLMConfig(ctx context.Context, userID uuid.UUID) (baseURL, token, model string, err error) {
	baseURL, err = a.settingsRepo.Get(ctx, "llm_api_url", userID)
	if err != nil {
		return "", "", "", err
	}
	if baseURL == "" {
		return "", "", "", fmt.Errorf("llm_api_url is not configured in settings")
	}

	token, err = a.settingsRepo.Get(ctx, "llm_api_token", userID)
	if err != nil {
		return "", "", "", err
	}

	model, _ = a.settingsRepo.Get(ctx, "llm_model", userID)
	if model == "" {
		model = defaultModel
	}

	return strings.TrimRight(baseURL, "/"), token, model, nil
}

func (a *Adapter) ParseBankStatementText(ctx context.Context, userID uuid.UUID, text string) (entity.BankStatement, error) {
	a.logger.Info("Starting AI bank statement parsing", "text_len", len(text), "user_id", userID)
	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_statement_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultStatementPromptTemplate
	}

	prompt := strings.ReplaceAll(promptTemplate, "{{TEXT}}", text)
	a.logger.Info("Sending statement extraction prompt to LLM", "prompt_len", len(prompt))

	rawResp, err := a.doRequest(ctx, userID, prompt)
	if err != nil {
		a.logger.Error("LLM request failed for statement extraction", "error", err)
		return entity.BankStatement{}, err
	}

	a.logger.Info("Received raw response from LLM for statement extraction", "raw_response", rawResp)
	rawResp = cleanJSONResponse(rawResp)

	// Temporary struct to handle string-based dates from the LLM
	type llmTx struct {
		BookingDate         string  `json:"booking_date"`
		Amount              float64 `json:"amount"`
		Description         string  `json:"description"`
		Reference           string  `json:"reference"`
		CounterpartyName    string  `json:"counterparty_name"`
		CounterpartyIban    string  `json:"counterparty_iban"`
		BankTransactionCode string  `json:"bank_transaction_code"`
		MandateReference    string  `json:"mandate_reference"`
	}

	var res struct {
		AccountHolder string  `json:"account_holder"`
		IBAN          string  `json:"iban"`
		Currency      string  `json:"currency"`
		StatementDate string  `json:"statement_date"`
		StatementNo   int     `json:"statement_no"`
		NewBalance    float64 `json:"new_balance"`
		Transactions  []llmTx `json:"transactions"`
	}

	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		a.logger.Error("failed to unmarshal AI bank statement", "err", err, "raw", rawResp)
		return entity.BankStatement{}, fmt.Errorf("llm adapter: parse json: %w", err)
	}

	// Map strings to time.Time gracefully
	stmtDate, _ := time.Parse("2006-01-02", res.StatementDate)
	if stmtDate.IsZero() {
		stmtDate = time.Now()
	}

	stmt := entity.BankStatement{
		AccountHolder: res.AccountHolder,
		IBAN:          res.IBAN,
		Currency:      res.Currency,
		StatementDate: stmtDate,
		StatementNo:   res.StatementNo,
		NewBalance:    res.NewBalance,
	}

	for _, tx := range res.Transactions {
		bDate, _ := time.Parse("2006-01-02", tx.BookingDate)
		if bDate.IsZero() {
			bDate = stmtDate
		}
		stmt.Transactions = append(stmt.Transactions, entity.Transaction{
			BookingDate:         bDate,
			Amount:              tx.Amount,
			Description:         tx.Description,
			Reference:           tx.Reference,
			CounterpartyName:    tx.CounterpartyName,
			CounterpartyIban:    tx.CounterpartyIban,
			BankTransactionCode: tx.BankTransactionCode,
			MandateReference:    tx.MandateReference,
		})
	}

	a.logger.Info("parsed bank statement", "raw", rawResp)
	return stmt, nil
}

func (a *Adapter) Categorize(ctx context.Context, userID uuid.UUID, req port.CategorizationRequest) (port.CategorizationResult, error) {
	a.logger.Info("Starting single categorization", "categories", req.Categories, "text_len", len(req.RawText), "user_id", userID)
	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_single_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultSinglePromptTemplate
	}

	prompt := strings.ReplaceAll(promptTemplate, "{{CATEGORIES}}", strings.Join(req.Categories, ", "))
	prompt = strings.ReplaceAll(prompt, "{{TEXT}}", req.RawText)
	a.logger.Info("Sending single categorization prompt to LLM", "prompt", prompt)

	rawResp, err := a.doRequest(ctx, userID, prompt)
	if err != nil {
		a.logger.Error("LLM request failed for single categorization", "error", err)
		return port.CategorizationResult{}, err
	}

	a.logger.Info("Received raw response from LLM for single categorization", "raw_response", rawResp)
	rawResp = cleanJSONResponse(rawResp)

	var res llmResponse
	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		a.logger.Error("failed to unmarshal single categorization", "err", err, "raw", rawResp)
		return port.CategorizationResult{}, fmt.Errorf("llm adapter: parse json: %w", err)
	}

	return port.CategorizationResult{
		CategoryName: strings.TrimSpace(res.CategoryName),
		VendorName:   strings.TrimSpace(res.VendorName),
		Amount:       res.Amount,
		Currency:     res.Currency,
		InvoiceDate:  res.InvoiceDate,
		Description:  res.Description,
	}, nil
}

func (a *Adapter) CategorizeBatch(ctx context.Context, userID uuid.UUID, txns []port.TransactionToCategorize, categories []string, examples []entity.CategorizationExample) ([]port.CategorizedTransaction, error) {
	a.logger.Info("Starting batch categorization", "txn_count", len(txns), "categories_count", len(categories), "examples_count", len(examples), "user_id", userID)
	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_batch_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultBatchPromptTemplate
	}

	txnsJSON, _ := json.Marshal(txns)
	examplesJSON, _ := json.Marshal(examples)

	prompt := strings.ReplaceAll(promptTemplate, "{{CATEGORIES}}", strings.Join(categories, ", "))
	prompt = strings.ReplaceAll(prompt, "{{EXAMPLES}}", string(examplesJSON))
	prompt = strings.ReplaceAll(prompt, "{{DATA}}", string(txnsJSON))

	a.logger.Info("Sending batch categorization prompt to LLM", "prompt", prompt)

	rawResp, err := a.doRequest(ctx, userID, prompt)
	if err != nil {
		a.logger.Error("LLM request failed for batch categorization", "error", err)
		return nil, err
	}

	a.logger.Info("Received raw response from LLM for batch categorization", "raw_response", rawResp)
	rawResp = cleanJSONResponse(rawResp)

	var results []port.CategorizedTransaction
	if err := json.Unmarshal([]byte(rawResp), &results); err != nil {
		var wrapped struct {
			Transactions []port.CategorizedTransaction `json:"transactions"`
			Data         []port.CategorizedTransaction `json:"data"`
			Results      []port.CategorizedTransaction `json:"results"`
		}

		if err := json.Unmarshal([]byte(rawResp), &wrapped); err != nil {
			a.logger.Error("failed to parse batch response", "err", err, "raw", rawResp)
			return nil, fmt.Errorf("llm adapter: could not parse batch JSON")
		}

		if len(wrapped.Transactions) > 0 {
			results = wrapped.Transactions
		} else if len(wrapped.Data) > 0 {
			results = wrapped.Data
		} else if len(wrapped.Results) > 0 {
			results = wrapped.Results
		} else {
			a.logger.Warn("llm returned an empty array or an unsupported wrapper schema", "raw", rawResp)
		}
	}

	for i := range results {
		results[i].Category = strings.TrimSpace(results[i].Category)
	}

	return results, nil
}

func (a *Adapter) doRequest(ctx context.Context, userID uuid.UUID, prompt string) (string, error) {
	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return "", err
	}

	if strings.Contains(baseURL, "googleapis.com") {
		return a.doGeminiRequest(ctx, baseURL, token, model, prompt)
	}
	return a.doOllamaRequest(ctx, baseURL, token, model, prompt)
}

func (a *Adapter) doGeminiRequest(ctx context.Context, url, token, model, prompt string) (string, error) {
	endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", url, model, token)

	payload, _ := json.Marshal(geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini api error: status %d", resp.StatusCode)
	}

	var gResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return "", err
	}

	if len(gResp.Candidates) > 0 && len(gResp.Candidates[0].Content.Parts) > 0 {
		return gResp.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("empty response from gemini")
}

func (a *Adapter) ParsePayslipText(ctx context.Context, userID uuid.UUID, text string) (entity.Payslip, error) {
	a.logger.Info("Starting AI payslip text parsing", "text_len", len(text), "user_id", userID)
	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_payslip_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultPayslipPromptTemplate
	}

	prompt := strings.ReplaceAll(promptTemplate, "{{TEXT}}", text)
	a.logger.Info("Sending payslip extraction prompt to LLM", "prompt_len", len(prompt))

	rawResp, err := a.doRequest(ctx, userID, prompt)
	if err != nil {
		a.logger.Error("LLM request failed for payslip extraction", "error", err)
		return entity.Payslip{}, err
	}

	a.logger.Info("Received raw response from LLM for payslip extraction", "raw_response", rawResp)
	return a.unmarshalPayslip(rawResp)
}

// ParsePayslipDocument satisfies the LLMPayslipParser interface used by the AI payslip parser.
// For Gemini, it sends the raw document bytes as a multimodal (inline_data) request.
// For Ollama (text-only), it falls back to the prompt-based text approach with a placeholder.
func (a *Adapter) ParsePayslipDocument(ctx context.Context, userID uuid.UUID, mimeType string, data []byte) (entity.Payslip, error) {
	a.logger.Info("Starting AI payslip document parsing (multimodal)", "mime_type", mimeType, "data_len", len(data), "user_id", userID)
	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return entity.Payslip{}, err
	}

	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_payslip_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultPayslipPromptTemplate
	}
	// Strip the {{TEXT}} placeholder for multimodal requests — the document IS the input
	systemPrompt := strings.ReplaceAll(promptTemplate, "\nSource Text:\n{{TEXT}}", "")

	var rawResp string
	if strings.Contains(baseURL, "googleapis.com") {
		a.logger.Info("Sending multimodal payslip extraction request to Gemini")
		// Gemini multimodal: send text prompt + inline binary document
		endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseURL, model, token)
		payload, _ := json.Marshal(geminiRequest{
			Contents: []geminiContent{{
				Parts: []geminiPart{
					{Text: systemPrompt},
					{InlineData: &geminiInlineData{
						MimeType: mimeType,
						Data:     base64.StdEncoding.EncodeToString(data),
					}},
				},
			}},
		})

		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			a.logger.Error("Gemini multimodal request failed", "error", err)
			return entity.Payslip{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			a.logger.Error("Gemini multimodal API error", "status", resp.StatusCode)
			return entity.Payslip{}, fmt.Errorf("gemini multimodal api error: status %d", resp.StatusCode)
		}

		var gResp geminiResponse
		if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
			return entity.Payslip{}, err
		}
		if len(gResp.Candidates) > 0 && len(gResp.Candidates[0].Content.Parts) > 0 {
			rawResp = gResp.Candidates[0].Content.Parts[0].Text
			a.logger.Info("Received raw response from Gemini for multimodal payslip extraction", "raw_response", rawResp)
		} else {
			return entity.Payslip{}, fmt.Errorf("empty multimodal response from gemini")
		}
	} else {
		a.logger.Info("Falling back to text-only payslip extraction for Ollama")
		// Ollama fallback: can't send binary blobs, so use the data as a UTF-8 text hint
		rawResp, err = a.doOllamaRequest(ctx, baseURL, token, model,
			strings.ReplaceAll(promptTemplate, "{{TEXT}}", "[Binary document — extract fields from PDF text above]"),
		)
		if err != nil {
			return entity.Payslip{}, err
		}
		a.logger.Info("Received raw response from Ollama for payslip extraction fallback", "raw_response", rawResp)
	}

	return a.unmarshalPayslip(rawResp)
}

// unmarshalPayslip is shared by ParsePayslipText and ParsePayslipDocument.
func (a *Adapter) unmarshalPayslip(rawResp string) (entity.Payslip, error) {
	rawResp = cleanJSONResponse(rawResp)

	var res struct {
		PeriodMonthNum   int     `json:"period_month_num"`
		PeriodYear       int     `json:"period_year"`
		EmployeeName     string  `json:"employee_name"`
		EmployerName     string  `json:"employer_name"` // Added EmployerName to the struct
		TaxClass         string  `json:"tax_class"`
		TaxID            string  `json:"tax_id"`
		GrossPay         float64 `json:"gross_pay"`
		NetPay           float64 `json:"net_pay"`
		PayoutAmount     float64 `json:"payout_amount"`
		CustomDeductions float64 `json:"custom_deductions"`
		Bonuses          []struct {
			Description string  `json:"description"`
			Amount      float64 `json:"amount"`
		} `json:"bonuses"`
	}

	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		a.logger.Error("failed to unmarshal AI payslip", "err", err, "raw", rawResp)
		return entity.Payslip{}, fmt.Errorf("llm adapter: parse payslip json: %w", err)
	}

	payslip := entity.Payslip{
		PeriodMonthNum:   res.PeriodMonthNum,
		PeriodYear:       res.PeriodYear,
		EmployeeName:     res.EmployeeName,
		EmployerName:     res.EmployerName, // Mapped the extracted value
		TaxClass:         res.TaxClass,
		TaxID:            res.TaxID,
		GrossPay:         res.GrossPay,
		NetPay:           res.NetPay,
		PayoutAmount:     res.PayoutAmount,
		CustomDeductions: res.CustomDeductions,
	}

	for _, b := range res.Bonuses {
		payslip.Bonuses = append(payslip.Bonuses, entity.Bonus{
			Description: b.Description,
			Amount:      b.Amount,
		})
	}

	a.logger.Info("successfully parsed payslip via LLM", "month", payslip.PeriodMonthNum, "net_pay", payslip.NetPay)
	return payslip, nil
}

func (a *Adapter) doOllamaRequest(ctx context.Context, url, token, model, prompt string) (string, error) {
	a.logger.Info("making request to ollama", "url", url, "model", model)
	payload, _ := json.Marshal(ollamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url+"/api/generate", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	slog.Info("ollama response received", "status", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama api error: status %d", resp.StatusCode)
	}

	var oResp ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	a.logger.Info("successfully generated ollama response", "response", oResp.Response)
	return oResp.Response, nil
}

func cleanJSONResponse(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")
	return strings.TrimSpace(input)
}
