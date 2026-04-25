package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"cogni-cash/internal/domain/entity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"log/slog"
	"cogni-cash/internal/domain/port"
)

type PayslipHandler struct {
	Logger *slog.Logger
	payslipRepo port.PayslipRepository
	payslipSvc port.PayslipUseCase
}

func NewPayslipHandler(Logger *slog.Logger, payslipRepo port.PayslipRepository, payslipSvc port.PayslipUseCase) *PayslipHandler {
	return &PayslipHandler{
		Logger: Logger,
		payslipRepo: payslipRepo,
		payslipSvc: payslipSvc,
	}
}

// allowedPayslipMIMETypes is the set of MIME types accepted by the payslip import endpoints.
// Image types are accepted because the AI payslip parser supports multimodal requests.
var allowedPayslipMIMETypes = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
}

// listPayslips handles GET /api/v1/payslips/
func (h *PayslipHandler) listPayslips(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter := entity.PayslipFilter{
		UserID:   userID,
		Employer: r.URL.Query().Get("employer"),
	}

	q := r.URL.Query()
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = offset
	}

	payslips, err := h.payslipSvc.GetAll(r.Context(), filter)
	if err != nil {
		h.Logger.Error("Failed to list payslips", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to fetch payslips")
		return
	}

	if payslips == nil {
		payslips = []entity.Payslip{} // Prevent null in JSON response
	}
	writeJSON(w, http.StatusOK, payslips)
}

// getPayslipSummary handles GET /api/v1/payslips/summary
func (h *PayslipHandler) getPayslipSummary(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.payslipSvc.GetSummary(r.Context(), userID)
	if err != nil {
		h.Logger.Error("Failed to get payslip summary", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to fetch payslip summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// getPayslip handles GET /api/v1/payslips/{id}
func (h *PayslipHandler) getPayslip(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing payslip ID")
		return
	}

	payslip, err := h.payslipSvc.GetByID(r.Context(), id, userID)
	if err != nil {
		h.Logger.Error("Failed to fetch payslip", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusNotFound, "Payslip not found")
		return
	}

	writeJSON(w, http.StatusOK, payslip)
}

// importPayslip handles POST /api/v1/payslips/import
func (h *PayslipHandler) importPayslip(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	const maxUpload = 10 << 20 // 10 MB hard cap
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "upload too large or could not parse form (max 10 MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing 'file' in form data")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to read file contents")
		return
	}

	mimeType := resolveMIME(header.Header.Get("Content-Type"), header.Filename)
	if mimeType == "application/octet-stream" {
		mimeType = "application/pdf"
	}
	if !allowedPayslipMIMETypes[mimeType] {
		writeError(w, http.StatusUnsupportedMediaType,
			fmt.Sprintf("unsupported file type %q — accepted types: PDF, JPEG, PNG, GIF, WEBP", mimeType))
		return
	}

	// Parse the Force AI flag
	useAI := r.FormValue("use_ai") == "true"

	// --- Extract Manual Overrides from Multipart Form ---
	overrides := &entity.Payslip{
		EmployerName: r.FormValue("employer_name"),
	}
	if val := r.FormValue("period_month_num"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			overrides.PeriodMonthNum = parsed
		}
	}
	if val := r.FormValue("period_year"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			overrides.PeriodYear = parsed
		}
	}
	if val := r.FormValue("tax_class"); val != "" {
		overrides.TaxClass = val
	}
	if val := r.FormValue("tax_id"); val != "" {
		overrides.TaxID = val
	}

	// Internationalized financial fields
	if val := r.FormValue("gross_pay"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			overrides.GrossPay = parsed
		}
	}
	if val := r.FormValue("net_pay"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			overrides.NetPay = parsed
		}
	}
	if val := r.FormValue("payout_amount"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			overrides.PayoutAmount = parsed
		}
	}
	if val := r.FormValue("custom_deductions"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			overrides.CustomDeductions = parsed
		}
	}
	// Bonuses overrides are sent as a JSON array string
	if val := r.FormValue("bonuses"); val != "" {
		var b []entity.Bonus
		if err := json.Unmarshal([]byte(val), &b); err == nil {
			overrides.Bonuses = b
		} else {
			h.Logger.Warn("Failed to parse bonuses override field", "error", err)
		}
	}
	// ----------------------------------------------------

	// Pass the fileBytes directly to the service layer (no more temp files)
	payslip, err := h.payslipSvc.Import(r.Context(), userID, header.Filename, mimeType, fileBytes, overrides, useAI)
	if err != nil {
		h.Logger.Error("Failed to import payslip", "error", err, "user_id", userID)
		if errors.Is(err, entity.ErrPayslipDuplicate) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusUnprocessableEntity, "Failed to parse payslip")
		return
	}

	writeJSON(w, http.StatusCreated, payslip)
}

// updatePayslip handles PUT /api/v1/payslips/{id}
func (h *PayslipHandler) updatePayslip(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing payslip ID")
		return
	}

	var p entity.Payslip
	contentType := r.Header.Get("Content-Type")

	// Check if the request is multipart (meaning a file was attached during edit)
	if len(contentType) >= 19 && contentType[:19] == "multipart/form-data" {
		const maxUpload = 10 << 20 // 10 MB hard cap
		r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
		if err := r.ParseMultipartForm(maxUpload); err != nil {
			writeError(w, http.StatusRequestEntityTooLarge, "upload too large or could not parse form (max 10 MB)")
			return
		}

		// The frontend sends the structured fields as a JSON string in the "data" form field
		dataStr := r.FormValue("data")
		if err := json.Unmarshal([]byte(dataStr), &p); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON in data field")
			return
		}

		// Extract the newly attached file
		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to read file contents")
				return
			}
			p.OriginalFileContent = fileBytes
			p.OriginalFileName = header.Filename
		}
	} else {
		// Standard JSON update (no file attached)
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
	}

	p.ID = id
	p.UserID = userID

	if err := h.payslipSvc.Update(r.Context(), &p); err != nil {
		h.Logger.Error("Failed to update payslip", "id", id, "error", err, "user_id", userID)
		if errors.Is(err, entity.ErrPayslipDuplicate) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update payslip")
		return
	}

	// Re-fetch to return the persisted state (including Bonuses and new file metadata)
	updated, err := h.payslipRepo.FindByID(r.Context(), id, userID)
	if err != nil {
		h.Logger.Error("Failed to fetch updated payslip", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to fetch updated payslip")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// deletePayslip handles DELETE /api/v1/payslips/{id}
func (h *PayslipHandler) deletePayslip(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing payslip ID")
		return
	}

	if err := h.payslipSvc.Delete(r.Context(), id, userID); err != nil {
		h.Logger.Error("Failed to delete payslip", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to delete payslip")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// downloadPayslipFile handles GET /api/v1/payslips/{id}/download
func (h *PayslipHandler) downloadPayslipFile(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing payslip ID")
		return
	}

	content, mimeType, filename, err := h.payslipSvc.GetOriginalFile(r.Context(), id, userID)

	if err != nil {
		h.Logger.Error("Failed to fetch original file", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusNotFound, "File not found")
		return
	}

	// JSON-imported payslips have no binary payload — signal this cleanly.
	if len(content) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if mimeType == "" {
		mimeType = "application/pdf"
	}

	// Set headers to force download with the correct filename
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))

	if _, err := w.Write(content); err != nil {
		h.Logger.Error("Failed to write file stream", "id", id, "error", err)
	}
}

// Response structures for the batch endpoint
type batchImportResponse struct {
	Successful []*entity.Payslip  `json:"successful"`
	Failed     []batchImportError `json:"failed"`
}

type batchImportError struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

// importPayslipsBatch handles POST /api/v1/payslips/import/batch
func (h *PayslipHandler) importPayslipsBatch(w http.ResponseWriter, r *http.Request) {
	const maxUpload = 50 << 20 // 50 MB hard cap for batch
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "upload too large or could not parse form (max 50 MB)")
		return
	}

	// Retrieve the array of files from the "files" form key
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "No 'files' found in form data")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	response := batchImportResponse{
		Successful: make([]*entity.Payslip, 0),
		Failed:     make([]batchImportError, 0),
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			response.Failed = append(response.Failed, batchImportError{
				Filename: fileHeader.Filename,
				Error:    "Failed to open uploaded file",
			})
			continue
		}

		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			response.Failed = append(response.Failed, batchImportError{
				Filename: fileHeader.Filename,
				Error:    "Failed to read file contents",
			})
			continue
		}

		mimeType := resolveMIME(fileHeader.Header.Get("Content-Type"), fileHeader.Filename)
		if mimeType == "application/octet-stream" {
			mimeType = "application/pdf"
		}
		if !allowedPayslipMIMETypes[mimeType] {
			response.Failed = append(response.Failed, batchImportError{
				Filename: fileHeader.Filename,
				Error:    fmt.Sprintf("unsupported file type %q — accepted: PDF, JPEG, PNG, GIF, WEBP", mimeType),
			})
			continue
		}

		// Call the service (passing nil for manual overrides in a batch context, and false for useAI)
		payslip, err := h.payslipSvc.Import(r.Context(), userID, fileHeader.Filename, mimeType, fileBytes, nil, false)

		if err != nil {
			// Provide a clean error message for duplicates
			errMsg := "Failed to parse payslip"
			if errors.Is(err, entity.ErrPayslipDuplicate) {
				errMsg = "Duplicate file (already imported)"
			}
			response.Failed = append(response.Failed, batchImportError{
				Filename: fileHeader.Filename,
				Error:    errMsg,
			})
			continue
		}

		response.Successful = append(response.Successful, payslip)
	}

	status := http.StatusOK
	if len(response.Successful) == 0 && len(response.Failed) > 0 {
		status = http.StatusUnprocessableEntity
	} else if len(response.Failed) > 0 {
		status = http.StatusMultiStatus // 207
	}

	writeJSON(w, status, response)
}
