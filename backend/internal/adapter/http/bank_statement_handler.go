package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"log/slog"
	"cogni-cash/internal/domain/port"
)

type BankStatementHandler struct {
	Logger *slog.Logger
	bankStatementSvc port.BankStatementUseCase
	bankStmtRepo port.BankStatementRepository
	forecastingSvc port.ForecastingUseCase
	settingsSvc port.SettingsUseCase
	transactionSvc port.TransactionUseCase
}

func NewBankStatementHandler(Logger *slog.Logger, bankStatementSvc port.BankStatementUseCase, bankStmtRepo port.BankStatementRepository, forecastingSvc port.ForecastingUseCase, settingsSvc port.SettingsUseCase, transactionSvc port.TransactionUseCase) *BankStatementHandler {
	return &BankStatementHandler{
		Logger: Logger,
		bankStatementSvc: bankStatementSvc,
		bankStmtRepo: bankStmtRepo,
		forecastingSvc: forecastingSvc,
		settingsSvc: settingsSvc,
		transactionSvc: transactionSvc,
	}
}

// allowedBankStatementMIMETypes is the set of MIME types accepted by the bank
// statement import endpoint. Image types are forwarded to the AI fallback parser
// which uses the multimodal (Gemini) path to extract statement data.
var allowedBankStatementMIMETypes = map[string]bool{
	"application/pdf":          true,
	"text/csv":                 true,
	"application/vnd.ms-excel": true, // .xls
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type updateTransactionCategoryRequest struct {
	CategoryID string `json:"category_id"`
}

type updateTransactionCategoriesBulkRequest struct {
	Hashes     []string `json:"hashes"`
	CategoryID string   `json:"category_id"`
}

func (h *BankStatementHandler) listBankStatements(w http.ResponseWriter, r *http.Request) {
	if h.bankStmtRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "bank statement repository not available")
		return
	}
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	summaries, err := h.bankStmtRepo.FindSummaries(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if summaries == nil {
		writeJSON(w, http.StatusOK, []entity.BankStatementSummary{})
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (h *BankStatementHandler) getBankStatement(w http.ResponseWriter, r *http.Request) {
	if h.bankStmtRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "bank statement repository not available")
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bank statement id")
		return
	}
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	stmt, err := h.bankStmtRepo.FindByID(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stmt)
}

func (h *BankStatementHandler) importBankStatement(w http.ResponseWriter, r *http.Request) {
	const maxUpload = 32 << 20 // 32 MB hard cap
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "upload too large or could not parse multipart form (max 32 MB)")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		files = r.MultipartForm.File["file"]
	}

	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "missing 'file' or 'files' field in form")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	useAI := r.FormValue("use_ai") == "true"

	// Capture the statement type from the frontend form (e.g., "giro", "extra_account", "credit_card")
	statementType := entity.StatementType(r.FormValue("statement_type"))

	type importResult struct {
		Filename string `json:"filename"`
		Status   string `json:"status"`
		Error    string `json:"error,omitempty"`
		ID       string `json:"id,omitempty"`
	}

	var results []importResult
	var importedCount int

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			results = append(results, importResult{Filename: fileHeader.Filename, Status: "error", Error: "could not open file"})
			continue
		}

		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			results = append(results, importResult{Filename: fileHeader.Filename, Status: "error", Error: "could not read file content"})
			continue
		}

		mimeType := resolveMIME(fileHeader.Header.Get("Content-Type"), fileHeader.Filename)
		if !allowedBankStatementMIMETypes[mimeType] {
			results = append(results, importResult{
				Filename: fileHeader.Filename,
				Status:   "error",
				Error:    fmt.Sprintf("unsupported file type %q — accepted: PDF, CSV, XLS, JPEG, PNG, GIF, WEBP", mimeType),
			})
			continue
		}

		// Images cannot be processed by the structured parsers — force AI path.
		effectiveUseAI := useAI || allowedBankStatementMIMETypes[mimeType] && isImageMIME(mimeType)

		// Pass the statementType down to the service layer.
		stmt, err := h.bankStatementSvc.ImportFromFile(r.Context(), userID, fileHeader.Filename, fileBytes, effectiveUseAI, statementType)

		if err != nil {
			if errors.Is(err, entity.ErrDuplicate) {
				results = append(results, importResult{Filename: fileHeader.Filename, Status: "duplicate"})
			} else {
				results = append(results, importResult{
					Filename: fileHeader.Filename,
					Status:   "error",
					Error:    mapImportError(err),
				})
			}
			continue
		}

		importedCount++
		results = append(results, importResult{Filename: fileHeader.Filename, Status: "imported", ID: stmt.ID.String()})
	}

	status := http.StatusOK
	if importedCount == 0 {
		allDuplicates := true
		for _, res := range results {
			if res.Status != "duplicate" {
				allDuplicates = false
				break
			}
		}
		if allDuplicates {
			status = http.StatusConflict
		} else {
			status = http.StatusUnprocessableEntity
		}
	} else if importedCount < len(files) {
		status = http.StatusMultiStatus
	}

	writeJSON(w, status, map[string]interface{}{
		"summary": map[string]int{
			"total":    len(files),
			"imported": importedCount,
		},
		"results": results,
	})
}

func (h *BankStatementHandler) deleteBankStatement(w http.ResponseWriter, r *http.Request) {
	if h.bankStatementSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "bank statement service not available")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bank statement id")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.bankStatementSvc.DeleteStatement(r.Context(), id, userID); err != nil {
		h.Logger.Error("Failed to delete bank statement", "error", err, "statement_id", id)
		writeError(w, http.StatusInternalServerError, "failed to delete bank statement")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BankStatementHandler) downloadBankStatementFile(w http.ResponseWriter, r *http.Request) {
	if h.bankStmtRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "bank statement repository not available")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bank statement id")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	stmt, err := h.bankStmtRepo.FindByID(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "bank statement not found")
		return
	}

	if len(stmt.OriginalFile) == 0 {
		writeError(w, http.StatusNotFound, "original file not available for this statement")
		return
	}

	contentType := http.DetectContentType(stmt.OriginalFile)

	// Manually check for Compound File Binary Format (BIFF8 XLS) magic bytes.
	// http.DetectContentType returns "application/octet-stream" for these files.
	if len(stmt.OriginalFile) >= 8 && bytes.HasPrefix(stmt.OriginalFile, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}) {
		contentType = "application/vnd.ms-excel"
	} else if contentType == "application/octet-stream" || contentType == "" {
		// Fall back to PDF which is the historical default for unrecognized binary data.
		contentType = "application/pdf"
	}
	ext := mimeToExt(contentType)
	filename := fmt.Sprintf("Statement_%s_%s%s", stmt.IBAN, stmt.StatementDate.Format("2006-01-02"), ext)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(stmt.OriginalFile)), 10))

	if _, err := w.Write(stmt.OriginalFile); err != nil {
		h.Logger.Error("Failed to write file to response", "error", err)
	}
}

func (h *BankStatementHandler) getTransactionAnalytics(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "transaction service unavailable")
		return
	}

	filter := h.parseTransactionFilter(r)
	analytics, err := h.transactionSvc.GetTransactionAnalytics(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch transaction analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

func (h *BankStatementHandler) listTransactions(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
		return
	}

	filter := h.parseTransactionFilter(r)
	txns, err := h.transactionSvc.ListTransactions(r.Context(), filter)
	if err != nil {
		h.Logger.Error("Failed to fetch transactions", "error", err, "filter", filter)
		writeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	if filter.IncludePredictions && h.forecastingSvc != nil {
		from := time.Now()
		if filter.FromDate != nil && filter.FromDate.After(from) {
			from = *filter.FromDate
		}
		to := from.AddDate(0, 0, 30)
		if filter.ToDate != nil {
			to = *filter.ToDate
		}

		forecast, err := h.forecastingSvc.GetCashFlowForecast(r.Context(), filter.UserID, from, to)
		if err == nil {
			for _, p := range forecast.Predictions {
				p.IsPrediction = true
				txns = append(txns, p.Transaction)
			}
			// Re-sort if we added predictions
			sort.Slice(txns, func(i, j int) bool {
				return txns[i].BookingDate.After(txns[j].BookingDate)
			})
		}
	}

	if txns == nil {
		writeJSON(w, http.StatusOK, []entity.Transaction{})
		return
	}
	writeJSON(w, http.StatusOK, txns)
}

func (h *BankStatementHandler) updateTransactionCategory(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "service not available")
		return
	}

	contentHash := chi.URLParam(r, "hash")
	if contentHash == "" {
		writeError(w, http.StatusBadRequest, "missing transaction hash")
		return
	}

	var req updateTransactionCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var catIDPtr *uuid.UUID
	if req.CategoryID != "" {
		parsedID, err := uuid.Parse(req.CategoryID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		catIDPtr = &parsedID
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.transactionSvc.UpdateCategory(r.Context(), contentHash, catIDPtr, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BankStatementHandler) updateTransactionCategoriesBulk(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "service not available")
		return
	}

	var req updateTransactionCategoriesBulkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Hashes) == 0 {
		writeError(w, http.StatusBadRequest, "missing transaction hashes")
		return
	}

	var catIDPtr *uuid.UUID
	if req.CategoryID != "" {
		parsedID, err := uuid.Parse(req.CategoryID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		catIDPtr = &parsedID
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.transactionSvc.UpdateCategoriesBulk(r.Context(), req.Hashes, catIDPtr, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BankStatementHandler) markTransactionReviewed(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "service not available")
		return
	}

	contentHash := chi.URLParam(r, "hash")
	if contentHash == "" {
		writeError(w, http.StatusBadRequest, "missing transaction hash")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.transactionSvc.MarkAsReviewed(r.Context(), contentHash, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BankStatementHandler) markTransactionsReviewedBulk(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "service not available")
		return
	}

	var payload struct {
		Hashes []string `json:"hashes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.transactionSvc.MarkAsReviewedBulk(r.Context(), payload.Hashes, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


func (h *BankStatementHandler) startAutoCategorize(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "transaction service not available")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	batchSize := 10
	if h.settingsSvc != nil {
		if settings, err := h.settingsSvc.GetAll(r.Context(), userID); err == nil {
			if bsStr, ok := settings["auto_categorization_batch_size"]; ok {
				if bs, err := strconv.Atoi(bsStr); err == nil && bs > 0 {
					batchSize = bs
				}
			}
		}
	}

	err := h.transactionSvc.StartAutoCategorizeAsync(r.Context(), userID, batchSize)
	if err != nil {
		if errors.Is(err, entity.ErrJobAlreadyRunning) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, entity.ErrNothingToCategorize) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		h.Logger.Error("Failed to start auto-categorization", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to start auto-categorization")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"message": "Batch categorization started"})
}

func (h *BankStatementHandler) getAutoCategorizeStatus(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "transaction service not available")
		return
	}
	status := h.transactionSvc.GetJobStatus()
	writeJSON(w, http.StatusOK, status)
}

func (h *BankStatementHandler) cancelAutoCategorize(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "transaction service not available")
		return
	}
	h.transactionSvc.CancelJob()
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cancellation requested"})
}

func (h *BankStatementHandler) parseTransactionFilter(r *http.Request) entity.TransactionFilter {
	q := r.URL.Query()
	f := entity.TransactionFilter{
		Type:   q.Get("type"),
		Search: q.Get("search"),
	}

	f.UserID = GetUserID(r.Context())

	if catID, err := uuid.Parse(q.Get("category_id")); err == nil {
		f.CategoryID = &catID
	}
	if sid, err := uuid.Parse(q.Get("statement_id")); err == nil {
		f.StatementID = &sid
	}
	if from, err := time.Parse("2006-01-02", q.Get("from")); err == nil {
		f.FromDate = &from
	}
	if to, err := time.Parse("2006-01-02", q.Get("to")); err == nil {
		f.ToDate = &to
	}
	if minAmt, err := strconv.ParseFloat(q.Get("min_amount"), 64); err == nil {
		f.MinAmount = &minAmt
	}
	if maxAmt, err := strconv.ParseFloat(q.Get("max_amount"), 64); err == nil {
		f.MaxAmount = &maxAmt
	}
	if q.Get("hide_reconciled") == "true" {
		isRec := false
		f.IsReconciled = &isRec
	} else if q.Get("hide_reconciled") == "false" {
		// If explicitly set to false, we want to see EVERYTHING.
		// Setting f.IsReconciled to nil achieves this in the repository.
		f.IsReconciled = nil
	}

	if q.Get("reviewed") == "true" {
		rev := true
		f.Reviewed = &rev
	} else if q.Get("reviewed") == "false" {
		rev := false
		f.Reviewed = &rev
	}

	if q.Get("include_predictions") == "true" {
		f.IncludePredictions = true
	}

	if q.Get("include_shared") == "true" {
		f.IncludeShared = true
	}

	// Add support for StatementType filtering
	if st := q.Get("statement_type"); st != "" {
		stType := entity.StatementType(st)
		f.StatementType = &stType
	}

	if subID, err := uuid.Parse(q.Get("subscription_id")); err == nil {
		f.SubscriptionID = &subID
	}

	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		f.Limit = limit
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		f.Offset = offset
	}

	return f
}

func mapImportError(err error) string {
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "unsupported format"):
		return "unsupported_format"
	case strings.Contains(errStr, "no suitable parser"):
		return "no_parser_found"
	case strings.Contains(errStr, "validation failed: invalid statement: missing iban"):
		return "missing_iban"
	case strings.Contains(errStr, "validation failed: invalid statement: missing statement date"):
		return "missing_date"
	case strings.Contains(errStr, "validation failed: invalid statement: zero transactions"):
		return "no_transactions"
	case strings.Contains(errStr, "corrupted") || strings.Contains(errStr, "failed to read"):
		return "corrupted_file"
	default:
		return "internal_error"
	}
}

func (h *BankStatementHandler) updateBankStatement(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bank statement id")
		return
	}

	var req struct {
		BankAccountID *uuid.UUID `json:"bank_account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.bankStatementSvc.UpdateStatementAccount(r.Context(), id, req.BankAccountID, userID); err != nil {
		h.Logger.Error("failed to update bank statement account", "error", err, "id", id)
		writeError(w, http.StatusInternalServerError, "failed to update statement account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
