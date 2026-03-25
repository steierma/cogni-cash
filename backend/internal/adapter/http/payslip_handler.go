package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/go-chi/chi/v5"
)

// listPayslips handles GET /api/v1/payslips/
func (h *Handler) listPayslips(w http.ResponseWriter, r *http.Request) {
	payslips, err := h.payslipRepo.FindAll(r.Context())
	if err != nil {
		h.Logger.Error("Failed to list payslips", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch payslips")
		return
	}

	if payslips == nil {
		payslips = []entity.Payslip{} // Prevent null in JSON response
	}
	writeJSON(w, http.StatusOK, payslips)
}

// getPayslip handles GET /api/v1/payslips/{id}
func (h *Handler) getPayslip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing payslip ID")
		return
	}

	payslip, err := h.payslipRepo.FindByID(r.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to fetch payslip", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "Payslip not found")
		return
	}

	writeJSON(w, http.StatusOK, payslip)
}

// importPayslip handles POST /api/v1/payslips/import
func (h *Handler) importPayslip(w http.ResponseWriter, r *http.Request) {
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

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/pdf" // Fallback assumption
	}

	// Parse the Force AI flag
	useAI := r.FormValue("use_ai") == "true"

	// --- Extract Manual Overrides from Multipart Form ---
	overrides := &entity.Payslip{
		EmployeeName: r.FormValue("employee_name"),
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

	// Create a temporary physical file for the PDF parser
	tempFile, err := os.CreateTemp("", "payslip-*.pdf")
	if err != nil {
		h.Logger.Error("Failed to create temporary file", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to process upload")
		return
	}
	// Ensure the file is deleted from the disk when the request finishes
	defer os.Remove(tempFile.Name())

	// Write the uploaded bytes to the temp file
	if _, err := tempFile.Write(fileBytes); err != nil {
		tempFile.Close()
		h.Logger.Error("Failed to write to temporary file", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to process upload")
		return
	}
	tempFile.Close() // Close it so the underlying parser can open it safely

	// Pass the actual physical temp file path, overrides, and useAI flag to the service layer
	payslip, err := h.payslipSvc.Import(r.Context(), tempFile.Name(), header.Filename, mimeType, fileBytes, overrides, useAI)
	if err != nil {
		h.Logger.Error("Failed to import payslip", "error", err)
		if errors.Is(err, service.ErrPayslipDuplicate) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusUnprocessableEntity, "Failed to parse payslip")
		return
	}

	writeJSON(w, http.StatusCreated, payslip)
}

// updatePayslip handles PUT /api/v1/payslips/{id}
func (h *Handler) updatePayslip(w http.ResponseWriter, r *http.Request) {
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
			mimeType := header.Header.Get("Content-Type")
			if mimeType == "" {
				mimeType = "application/pdf"
			}
			p.OriginalFileMime = mimeType
			p.OriginalFileSize = int64(len(fileBytes))
		}
	} else {
		// Standard JSON update (no file attached)
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
	}

	p.ID = id

	if err := h.payslipSvc.Update(r.Context(), &p); err != nil {
		h.Logger.Error("Failed to update payslip", "id", id, "error", err)
		if errors.Is(err, service.ErrPayslipDuplicate) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update payslip")
		return
	}

	// Re-fetch to return the persisted state (including Bonuses and new file metadata)
	updated, err := h.payslipRepo.FindByID(r.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to fetch updated payslip", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch updated payslip")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// deletePayslip handles DELETE /api/v1/payslips/{id}
func (h *Handler) deletePayslip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing payslip ID")
		return
	}

	if err := h.payslipSvc.Delete(r.Context(), id); err != nil {
		h.Logger.Error("Failed to delete payslip", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to delete payslip")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// downloadPayslipFile handles GET /api/v1/payslips/{id}/download
func (h *Handler) downloadPayslipFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing payslip ID")
		return
	}

	content, mimeType, filename, err := h.payslipRepo.GetOriginalFile(r.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to fetch original file", "id", id, "error", err)
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
func (h *Handler) importPayslipsBatch(w http.ResponseWriter, r *http.Request) {
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

		mimeType := fileHeader.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "application/pdf"
		}

		// Create a temporary physical file for the PDF parser
		tempFile, err := os.CreateTemp("", "payslip-batch-*.pdf")
		if err != nil {
			h.Logger.Error("Failed to create temporary file in batch", "error", err)
			response.Failed = append(response.Failed, batchImportError{
				Filename: fileHeader.Filename,
				Error:    "Internal server error processing file",
			})
			continue
		}

		// Write the uploaded bytes to the temp file
		if _, err := tempFile.Write(fileBytes); err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			response.Failed = append(response.Failed, batchImportError{
				Filename: fileHeader.Filename,
				Error:    "Failed to write to temporary file",
			})
			continue
		}
		tempFile.Close()

		// Call the service (passing nil for manual overrides in a batch context, and false for useAI)
		payslip, err := h.payslipSvc.Import(r.Context(), tempFile.Name(), fileHeader.Filename, mimeType, fileBytes, nil, false)

		// Clean up the temp file immediately after the service is done
		os.Remove(tempFile.Name())

		if err != nil {
			// Provide a clean error message for duplicates
			errMsg := "Failed to parse payslip"
			if errors.Is(err, service.ErrPayslipDuplicate) {
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
