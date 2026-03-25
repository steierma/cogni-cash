package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type updateTransactionCategoryRequest struct {
	CategoryID string `json:"category_id"`
}

func (h *Handler) listBankStatements(w http.ResponseWriter, r *http.Request) {
	if h.bankStmtRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "bank statement repository not available")
		return
	}
	summaries, err := h.bankStmtRepo.FindSummaries(r.Context())
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

func (h *Handler) getBankStatement(w http.ResponseWriter, r *http.Request) {
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
	stmt, err := h.bankStmtRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stmt)
}

func (h *Handler) importBankStatement(w http.ResponseWriter, r *http.Request) {
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

		tmp, err := os.CreateTemp("", "bank-statement-*"+filepath.Ext(fileHeader.Filename))
		if err != nil {
			file.Close()
			results = append(results, importResult{Filename: fileHeader.Filename, Status: "error", Error: "could not create temp file"})
			continue
		}

		_, copyErr := io.Copy(tmp, file)
		tmp.Close()
		file.Close()

		if copyErr != nil {
			os.Remove(tmp.Name())
			results = append(results, importResult{Filename: fileHeader.Filename, Status: "error", Error: "could not read file content"})
			continue
		}

		// Pass the statementType down to the service layer.
		stmt, err := h.bankStatementSvc.ImportFromFile(r.Context(), tmp.Name(), useAI, statementType)
		os.Remove(tmp.Name())

		if err != nil {
			if errors.Is(err, service.ErrDuplicate) {
				results = append(results, importResult{Filename: fileHeader.Filename, Status: "duplicate"})
			} else {
				results = append(results, importResult{Filename: fileHeader.Filename, Status: "error", Error: err.Error()})
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

func (h *Handler) deleteBankStatement(w http.ResponseWriter, r *http.Request) {
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

	if err := h.bankStatementSvc.DeleteStatement(r.Context(), id); err != nil {
		h.Logger.Error("Failed to delete bank statement", "error", err, "statement_id", id)
		writeError(w, http.StatusInternalServerError, "failed to delete bank statement")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) downloadBankStatementFile(w http.ResponseWriter, r *http.Request) {
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

	stmt, err := h.bankStmtRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bank statement not found")
		return
	}

	if len(stmt.OriginalFile) == 0 {
		writeError(w, http.StatusNotFound, "original file not available for this statement")
		return
	}

	ext := strings.ToLower(filepath.Ext(stmt.SourceFile))
	contentType := "application/octet-stream"
	switch ext {
	case ".pdf":
		contentType = "application/pdf"
	case ".csv":
		contentType = "text/csv"
	case ".xls":
		contentType = "application/vnd.ms-excel"
	case ".xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", stmt.SourceFile))
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(stmt.OriginalFile)), 10))

	if _, err := w.Write(stmt.OriginalFile); err != nil {
		h.Logger.Error("Failed to write file to response", "error", err)
	}
}

func (h *Handler) getTransactionAnalytics(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "transaction service unavailable")
		return
	}

	filter := parseTransactionFilter(r)
	analytics, err := h.transactionSvc.GetTransactionAnalytics(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch transaction analytics")
		return
	}

	writeJSON(w, http.StatusOK, analytics)
}

func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	if h.bankStmtRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "repository unavailable")
		return
	}

	filter := parseTransactionFilter(r)
	txns, err := h.bankStmtRepo.FindTransactions(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	if txns == nil {
		writeJSON(w, http.StatusOK, []entity.Transaction{})
		return
	}
	writeJSON(w, http.StatusOK, txns)
}

func (h *Handler) updateTransactionCategory(w http.ResponseWriter, r *http.Request) {
	if h.bankStmtRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "repository not available")
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

	if err := h.bankStmtRepo.UpdateTransactionCategory(r.Context(), contentHash, catIDPtr); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) startAutoCategorize(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "transaction service not available")
		return
	}

	batchSize := 10
	if h.settingsSvc != nil {
		if settings, err := h.settingsSvc.GetAll(r.Context()); err == nil {
			if bsStr, ok := settings["auto_categorization_batch_size"]; ok {
				if bs, err := strconv.Atoi(bsStr); err == nil && bs > 0 {
					batchSize = bs
				}
			}
		}
	}

	err := h.transactionSvc.StartAutoCategorizeAsync(r.Context(), batchSize)
	if err != nil {
		if errors.Is(err, service.ErrJobAlreadyRunning) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, service.ErrNothingToCategorize) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		h.Logger.Error("Failed to start auto-categorization", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to start auto-categorization")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"message": "Batch categorization started"})
}

func (h *Handler) getAutoCategorizeStatus(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "transaction service not available")
		return
	}
	status := h.transactionSvc.GetJobStatus()
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) cancelAutoCategorize(w http.ResponseWriter, r *http.Request) {
	if h.transactionSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "transaction service not available")
		return
	}
	h.transactionSvc.CancelJob()
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cancellation requested"})
}

func parseTransactionFilter(r *http.Request) entity.TransactionFilter {
	q := r.URL.Query()
	f := entity.TransactionFilter{
		Type:   q.Get("type"),
		Search: q.Get("search"),
	}

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
	}
	// Add support for StatementType filtering
	if st := q.Get("statement_type"); st != "" {
		stType := entity.StatementType(st)
		f.StatementType = &stType
	}

	return f
}
