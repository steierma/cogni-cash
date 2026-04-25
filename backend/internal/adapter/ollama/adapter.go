package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const defaultModel = "gemini-1.5-flash"

// --- Google AI Types ---

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

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
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Stream bool     `json:"stream"`
	Format string   `json:"format"`
	Images []string `json:"images,omitempty"`
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
	extractor    port.InvoiceParser // Injected to handle internal fallback for PDFs
	client       *http.Client
	logger       *slog.Logger
}

func NewAdapter(settingsRepo port.SettingsRepository, extractor port.InvoiceParser, logger *slog.Logger) *Adapter {
	if logger == nil {
		logger = slog.Default()
	}
	return &Adapter{
		settingsRepo: settingsRepo,
		extractor:    extractor,
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

// ─── CORE PROCESSING ENGINE (The Internal Fallback Logic) ────────────────────

// processFileWithAI determines the capabilities of the configured model and routes the file appropriately.
// It acts as the fallback mechanism if a non-multimodal model is given a PDF.
func (a *Adapter) processFileWithAI(ctx context.Context, userID uuid.UUID, baseURL, token, model, mimeType string, fileBytes []byte, promptTemplate string) (string, error) {
	imageTypes := map[string]bool{
		"image/jpeg": true, "image/jpg": true, "image/png": true, "image/gif": true, "image/webp": true,
	}

	isGemini := strings.Contains(baseURL, "googleapis.com")
	isPDF := mimeType == "application/pdf"
	isImage := imageTypes[mimeType]

	// Path 1: Native Multimodal Processing
	if isImage || (isGemini && isPDF) {
		a.logger.Info("Routing to native multimodal endpoint")
		systemPrompt := strings.ReplaceAll(promptTemplate, "\nTEXT:\n{{TEXT}}", "")
		systemPrompt = strings.ReplaceAll(systemPrompt, "\nSource Text:\n{{TEXT}}", "")
		systemPrompt = strings.ReplaceAll(systemPrompt, "{{TEXT}}", "")

		if isGemini {
			return a.doGeminiMultimodalRequest(ctx, baseURL, token, model, systemPrompt, mimeType, fileBytes)
		}
		return a.doOllamaRequest(ctx, baseURL, token, model, systemPrompt, []string{base64.StdEncoding.EncodeToString(fileBytes)})
	}

	// Path 2: Internal Fallback for Text and PDFs without native multimodal support
	a.logger.Info("Routing to text-only endpoint (triggering internal extraction if needed)")

	var rawText string
	if isPDF {
		if a.extractor == nil {
			return "", fmt.Errorf("cannot process PDF: multimodal not supported by provider and text extractor is not configured")
		}
		extracted, err := a.extractor.Extract(ctx, userID, fileBytes, mimeType)
		if err != nil || extracted == "" {
			return "", fmt.Errorf("internal fallback text extraction failed: %w", err)
		}
		rawText = extracted
	} else {
		// Plain text files
		rawText = string(fileBytes)
	}

	finalPrompt := strings.ReplaceAll(promptTemplate, "{{TEXT}}", rawText)
	return a.doTextOnlyRequest(ctx, baseURL, token, model, finalPrompt)
}

func (a *Adapter) doTextOnlyRequest(ctx context.Context, baseURL, token, model, prompt string) (string, error) {
	if strings.Contains(baseURL, "googleapis.com") {
		return a.doAIRequest(ctx, baseURL, token, model, prompt)
	}
	return a.doOllamaRequest(ctx, baseURL, token, model, prompt, nil)
}

// ─── BANK STATEMENT PARSING ──────────────────────────────────────────────────

func (a *Adapter) ParseBankStatement(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.BankStatement, error) {
	a.logger.Info("AI bank statement parsing initiated", "file", fileName, "mime_type", mimeType, "user_id", userID)

	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return entity.BankStatement{}, err
	}

	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_statement_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultStatementPromptTemplate
	}

	fallbackCurrency, _ := a.settingsRepo.Get(ctx, "currency", userID)
	if strings.TrimSpace(fallbackCurrency) == "" {
		fallbackCurrency = "EUR"
	}

	promptTemplate = strings.ReplaceAll(promptTemplate, "{{DEFAULT_CURRENCY}}", fallbackCurrency)

	rawResp, err := a.processFileWithAI(ctx, userID, baseURL, token, model, mimeType, fileBytes, promptTemplate)
	if err != nil {
		return entity.BankStatement{}, err
	}

	return a.unmarshalBankStatement(rawResp, fallbackCurrency)
}

func (a *Adapter) unmarshalBankStatement(rawResp string, fallbackCurrency string) (entity.BankStatement, error) {
	rawResp = cleanJSONResponse(rawResp)

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

	stmtDate, _ := time.Parse("2006-01-02", res.StatementDate)
	if stmtDate.IsZero() {
		stmtDate = time.Now()
	}

	stmt := entity.BankStatement{
		AccountHolder: res.AccountHolder,
		IBAN:          res.IBAN,
		Currency:      a.normalizeCurrency(res.Currency, fallbackCurrency),
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
			Currency:            stmt.Currency, // Inherit from statement
		})
	}

	return stmt, nil
}

// ─── PAYSLIP PARSING ─────────────────────────────────────────────────────────

func (a *Adapter) ParsePayslip(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.Payslip, error) {
	a.logger.Info("AI payslip parsing initiated", "file", fileName, "mime_type", mimeType, "user_id", userID)

	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return entity.Payslip{}, err
	}

	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_payslip_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultPayslipPromptTemplate
	}

	rawResp, err := a.processFileWithAI(ctx, userID, baseURL, token, model, mimeType, fileBytes, promptTemplate)
	if err != nil {
		return entity.Payslip{}, err
	}

	return a.unmarshalPayslip(rawResp)
}

func (a *Adapter) unmarshalPayslip(rawResp string) (entity.Payslip, error) {
	rawResp = cleanJSONResponse(rawResp)

	var res struct {
		PeriodMonthNum   int     `json:"period_month_num"`
		PeriodYear       int     `json:"period_year"`
		EmployerName     string  `json:"employer_name"`
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
		EmployerName:     res.EmployerName,
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

	return payslip, nil
}

// ─── INVOICE & TRANSACTION PARSING ───────────────────────────────────────────

func (a *Adapter) CategorizeInvoice(ctx context.Context, userID uuid.UUID, req port.CategorizationRequest) (port.InvoiceCategorizationResult, error) {
	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_single_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultSinglePromptTemplate
	}

	fallbackCurrency, _ := a.settingsRepo.Get(ctx, "currency", userID)
	if strings.TrimSpace(fallbackCurrency) == "" {
		fallbackCurrency = "EUR"
	}

	prompt := strings.ReplaceAll(promptTemplate, "{{CATEGORIES}}", strings.Join(req.Categories, ", "))
	prompt = strings.ReplaceAll(prompt, "{{TEXT}}", req.RawText)
	prompt = strings.ReplaceAll(prompt, "{{DEFAULT_CURRENCY}}", fallbackCurrency)

	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return port.InvoiceCategorizationResult{}, err
	}

	rawResp, err := a.doTextOnlyRequest(ctx, baseURL, token, model, prompt)
	if err != nil {
		a.logger.Error("LLM request failed for single categorization", "error", err)
		return port.InvoiceCategorizationResult{}, err
	}

	rawResp = cleanJSONResponse(rawResp)

	var res llmResponse
	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		a.logger.Error("failed to unmarshal single categorization", "err", err, "raw", rawResp)
		return port.InvoiceCategorizationResult{}, fmt.Errorf("llm adapter: parse json: %w", err)
	}

	return port.InvoiceCategorizationResult{
		InvoiceName: strings.TrimSpace(res.CategoryName),
		VendorName:  strings.TrimSpace(res.VendorName),
		Amount:      res.Amount,
		Currency:    a.normalizeCurrency(res.Currency, fallbackCurrency),
		InvoiceDate: res.InvoiceDate,
		Description: res.Description,
	}, nil
}

func (a *Adapter) normalizeCurrency(extracted, fallback string) string {
	extracted = strings.TrimSpace(strings.ToUpper(extracted))
	if len(extracted) == 3 {
		return extracted
	}
	return strings.ToUpper(fallback)
}

func (a *Adapter) CategorizeInvoiceImage(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, imageBytes []byte, categories []string) (port.InvoiceCategorizationResult, error) {
	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return port.InvoiceCategorizationResult{}, err
	}

	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_single_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultSinglePromptTemplate
	}

	fallbackCurrency, _ := a.settingsRepo.Get(ctx, "currency", userID)
	if strings.TrimSpace(fallbackCurrency) == "" {
		fallbackCurrency = "EUR"
	}

	catStr := strings.Join(categories, ", ")
	systemPrompt := strings.ReplaceAll(promptTemplate, "{{CATEGORIES}}", catStr)
	systemPrompt = strings.ReplaceAll(systemPrompt, "\n\nTEXT:\n{{TEXT}}", "")
	systemPrompt = strings.ReplaceAll(systemPrompt, "\nTEXT:\n{{TEXT}}", "")
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{TEXT}}", "")
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{DEFAULT_CURRENCY}}", fallbackCurrency)

	var rawResp string
	isGemini := strings.Contains(baseURL, "googleapis.com")

	if isGemini {
		rawResp, err = a.doGeminiMultimodalRequest(ctx, baseURL, token, model, systemPrompt, mimeType, imageBytes)
	} else {
		images := []string{base64.StdEncoding.EncodeToString(imageBytes)}
		rawResp, err = a.doOllamaRequest(ctx, baseURL, token, model, systemPrompt, images)
	}

	if err != nil {
		return port.InvoiceCategorizationResult{}, err
	}

	rawResp = cleanJSONResponse(rawResp)

	var res llmResponse
	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		return port.InvoiceCategorizationResult{}, fmt.Errorf("llm adapter: parse image categorization json: %w", err)
	}

	return port.InvoiceCategorizationResult{
		InvoiceName: strings.TrimSpace(res.CategoryName),
		VendorName:  strings.TrimSpace(res.VendorName),
		Amount:      res.Amount,
		Currency:    a.normalizeCurrency(res.Currency, fallbackCurrency),
		InvoiceDate: res.InvoiceDate,
		Description: res.Description,
	}, nil
}

func (a *Adapter) CategorizeTransactionsBatch(ctx context.Context, userID uuid.UUID, txns []port.TransactionToCategorize, categories []string, examples []entity.CategorizationExample) ([]port.CategorizedTransaction, error) {
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

	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return nil, err
	}

	rawResp, err := a.doTextOnlyRequest(ctx, baseURL, token, model, prompt)
	if err != nil {
		a.logger.Error("LLM request failed for batch categorization", "error", err)
		return nil, err
	}

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

// ─── DOCUMENT VAULT PARSING ──────────────────────────────────────────────────

func (a *Adapter) ClassifyAndExtract(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.DocumentType, map[string]interface{}, string, error) {
	a.logger.Info("AI document classification and extraction initiated", "file", fileName, "mime_type", mimeType, "user_id", userID)

	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return entity.DocTypeOther, nil, "", err
	}

	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_document_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultDocumentPromptTemplate
	}

	// 1. Extract text for search (if PDF/Text)
	var extractedText string
	if a.extractor != nil {
		// This handles PDFs
		extractedText, _ = a.extractor.Extract(ctx, userID, fileBytes, mimeType)
	}

	isPDF := strings.Contains(mimeType, "application/pdf")
	isImage := strings.HasPrefix(mimeType, "image/")

	if extractedText == "" && !isPDF && !isImage {
		extractedText = string(fileBytes)
	}

	// 2. Call LLM for classification and metadata
	rawResp, err := a.processFileWithAI(ctx, userID, baseURL, token, model, mimeType, fileBytes, promptTemplate)
	if err != nil {
		return entity.DocTypeOther, nil, extractedText, err
	}

	return a.unmarshalDocumentClassification(rawResp, extractedText)
}

func (a *Adapter) VerifySubscriptionSuggestion(ctx context.Context, userID uuid.UUID, merchantName string, amount float64, currency string, billingCycle string) (bool, error) {
	a.logger.Info("AI subscription verification initiated", "merchant", merchantName, "user_id", userID)

	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return false, err
	}

	prompt := defaultVerificationPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{MERCHANT}}", merchantName)
	prompt = strings.ReplaceAll(prompt, "{{AMOUNT}}", fmt.Sprintf("%.2f", amount))
	prompt = strings.ReplaceAll(prompt, "{{CURRENCY}}", currency)
	prompt = strings.ReplaceAll(prompt, "{{CYCLE}}", billingCycle)

	rawResp, err := a.doTextOnlyRequest(ctx, baseURL, token, model, prompt)
	if err != nil {
		return false, err
	}

	rawResp = cleanJSONResponse(rawResp)

	var res struct {
		IsSubscription bool `json:"is_subscription"`
	}
	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		a.logger.Error("failed to unmarshal AI subscription verification", "err", err, "raw", rawResp)
		// Default to true if AI fails, to not miss potential subscriptions (conservative)
		// Or false to be "stricter" as requested. Let's go with false for strictness.
		return false, fmt.Errorf("llm adapter: parse verification json: %w", err)
	}

	return res.IsSubscription, nil
}

func (a *Adapter) EnrichSubscription(ctx context.Context, userID uuid.UUID, merchantName string, transactionDescriptions []string, language string) (port.SubscriptionEnrichmentResult, error) {
	a.logger.Info("AI subscription enrichment initiated", "merchant", merchantName, "user_id", userID, "language", language)

	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return port.SubscriptionEnrichmentResult{}, err
	}

	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_subscription_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultSubscriptionPromptTemplate
	}

	// Use full language names for better AI context
	langName := language
	if strings.ToLower(language) == "de" {
		langName = "deutsch"
	} else if strings.ToLower(language) == "en" {
		langName = "english"
	} else if strings.ToLower(language) == "es" {
		langName = "spanish"
	} else if strings.ToLower(language) == "fr" {
		langName = "french"
	}

	txText := strings.Join(transactionDescriptions, "\n- ")
	prompt := strings.ReplaceAll(promptTemplate, "{{MERCHANT}}", merchantName)
	prompt = strings.ReplaceAll(prompt, "{{TRANSACTIONS}}", txText)
	prompt = strings.ReplaceAll(prompt, "{{LANGUAGE}}", langName)

	rawResp, err := a.doTextOnlyRequest(ctx, baseURL, token, model, prompt)
	if err != nil {
		return port.SubscriptionEnrichmentResult{}, err
	}

	rawResp = cleanJSONResponse(rawResp)

	var res port.SubscriptionEnrichmentResult
	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		a.logger.Error("failed to unmarshal AI subscription enrichment", "err", err, "raw", rawResp)
		return port.SubscriptionEnrichmentResult{}, fmt.Errorf("llm adapter: parse json: %w", err)
	}

	return res, nil
}

func (a *Adapter) GenerateCancellationLetter(ctx context.Context, userID uuid.UUID, req port.CancellationLetterRequest) (port.CancellationLetterResult, error) {
	a.logger.Info("AI cancellation letter generation initiated", "merchant", req.MerchantName, "user_id", userID)

	baseURL, token, model, err := a.getLLMConfig(ctx, userID)
	if err != nil {
		return port.CancellationLetterResult{}, err
	}

	promptTemplate, _ := a.settingsRepo.Get(ctx, "llm_cancellation_prompt", userID)
	if strings.TrimSpace(promptTemplate) == "" {
		promptTemplate = defaultCancellationLetterPromptTemplate
	}

	prompt := strings.ReplaceAll(promptTemplate, "{{USER_NAME}}", req.UserFullName)
	prompt = strings.ReplaceAll(prompt, "{{USER_EMAIL}}", req.UserEmail)
	prompt = strings.ReplaceAll(prompt, "{{MERCHANT}}", req.MerchantName)
	prompt = strings.ReplaceAll(prompt, "{{CUSTOMER_NUMBER}}", req.CustomerNumber)
	prompt = strings.ReplaceAll(prompt, "{{END_DATE}}", req.ContractEndDate)
	prompt = strings.ReplaceAll(prompt, "{{NOTICE_PERIOD}}", fmt.Sprintf("%d", req.NoticePeriodDays))
	prompt = strings.ReplaceAll(prompt, "{{LANGUAGE}}", req.Language)

	rawResp, err := a.doTextOnlyRequest(ctx, baseURL, token, model, prompt)
	if err != nil {
		return port.CancellationLetterResult{}, err
	}

	rawResp = cleanJSONResponse(rawResp)

	var res port.CancellationLetterResult
	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		a.logger.Error("failed to unmarshal AI cancellation letter", "err", err, "raw", rawResp)
		return port.CancellationLetterResult{}, fmt.Errorf("llm adapter: parse json: %w", err)
	}

	return res, nil
}

func (a *Adapter) unmarshalDocumentClassification(rawResp string, extractedText string) (entity.DocumentType, map[string]interface{}, string, error) {
	rawResp = cleanJSONResponse(rawResp)

	var res struct {
		DocumentType string                 `json:"document_type"`
		Metadata     map[string]interface{} `json:"metadata"`
		Summary      string                 `json:"summary"`
	}

	if err := json.Unmarshal([]byte(rawResp), &res); err != nil {
		a.logger.Error("failed to unmarshal AI document classification", "err", err, "raw", rawResp)
		return entity.DocTypeOther, nil, extractedText, fmt.Errorf("llm adapter: parse json: %w", err)
	}

	docType := entity.DocumentType(strings.ToLower(res.DocumentType))
	switch docType {
	case entity.DocTypeTaxCertificate, entity.DocTypeReceipt, entity.DocTypeContract, entity.DocTypeOther:
		// valid
	default:
		docType = entity.DocTypeOther
	}

	if res.Metadata == nil {
		res.Metadata = make(map[string]interface{})
	}
	if res.Summary != "" {
		res.Metadata["summary"] = res.Summary
	}

	return docType, res.Metadata, extractedText, nil
}

// ─── HTTP REQUEST HELPERS ────────────────────────────────────────────────────

func (a *Adapter) doGeminiMultimodalRequest(ctx context.Context, url, token, model, prompt, mimeType string, fileBytes []byte) (string, error) {
	endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", url, model, token)
	payload, err := json.Marshal(geminiRequest{
		Contents: []geminiContent{{
			Parts: []geminiPart{
				{Text: prompt},
				{InlineData: &geminiInlineData{
					MimeType: mimeType,
					Data:     base64.StdEncoding.EncodeToString(fileBytes),
				}},
			},
		}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal gemini payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini multimodal api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &gResp); err != nil {
		return "", fmt.Errorf("failed to decode gemini response: %w", err)
	}
	if len(gResp.Candidates) > 0 && len(gResp.Candidates[0].Content.Parts) > 0 {
		return gResp.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("empty response from gemini")
}

func (a *Adapter) doAIRequest(ctx context.Context, url, token, model, prompt string) (string, error) {
	endpoint := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", url, model, token)
	payload, err := json.Marshal(geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal gemini payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &gResp); err != nil {
		return "", fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if len(gResp.Candidates) > 0 && len(gResp.Candidates[0].Content.Parts) > 0 {
		return gResp.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("empty response from gemini")
}

func (a *Adapter) doOllamaRequest(ctx context.Context, url, token, model, prompt string, images []string) (string, error) {
	payload, err := json.Marshal(ollamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
		Images: images,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama payload: %w", err)
	}

	fullURL := url + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var oResp ollamaGenerateResponse
	if err := json.Unmarshal(bodyBytes, &oResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}
	return oResp.Response, nil
}

func cleanJSONResponse(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")
	return strings.TrimSpace(input)
}
