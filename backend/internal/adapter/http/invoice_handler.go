package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ── request / response types ─────────────────────────────────────────────────

type categorizeDocumentRequest struct {
	RawText string `json:"raw_text"`
}

type updateInvoiceRequest struct {
	VendorName  string     `json:"vendor_name"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	IssuedAt    *time.Time `json:"issued_at"`
	Description *string    `json:"description"` // pointer so empty-string clears the field
	RawText     string     `json:"raw_text"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// GET /api/v1/invoices/
func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.invoiceSvc.GetAll(r.Context())
	if err != nil {
		h.Logger.Error("Failed to list invoices", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if invoices == nil {
		invoices = []entity.Invoice{}
	}
	writeJSON(w, http.StatusOK, invoices)
}

// GET /api/v1/invoices/{id}
func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}
	invoice, err := h.invoiceSvc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrInvoiceNotFound) {
			writeError(w, http.StatusNotFound, "invoice not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, invoice)
}

// POST /api/v1/invoices/import   (multipart/form-data, field "file")
func (h *Handler) importInvoice(w http.ResponseWriter, r *http.Request) {
	const maxUpload = 32 << 20 // 32 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "upload too large or could not parse form (max 32 MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field in form")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/pdf"
	}

	// Optional: caller-specified category override
	var categoryID *uuid.UUID
	if catStr := r.FormValue("category_id"); catStr != "" {
		parsed, err := uuid.Parse(catStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		categoryID = &parsed
	}

	invoice, err := h.invoiceSvc.ImportFromFile(r.Context(), "", header.Filename, mimeType, fileBytes, categoryID)
	if err != nil {
		if errors.Is(err, entity.ErrInvoiceDuplicate) {
			writeError(w, http.StatusConflict, "invoice already imported (duplicate)")
			return
		}
		h.Logger.Error("Failed to import invoice", "error", err)
		writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("import failed: %s", err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, invoice)
}

// PUT /api/v1/invoices/{id}
func (h *Handler) updateInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	var req updateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Fetch current record to preserve immutable fields
	existing, err := h.invoiceSvc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrInvoiceNotFound) {
			writeError(w, http.StatusNotFound, "invoice not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated := existing
	if req.VendorName != "" {
		updated.Vendor.Name = req.VendorName
	}
	// CategoryID: JSON null decodes to nil (*uuid.UUID), which correctly clears it.
	// We always overwrite so the caller can unset a category.
	updated.CategoryID = req.CategoryID
	if req.Amount != 0 {
		updated.Amount = req.Amount
	}
	if req.Currency != "" {
		updated.Currency = req.Currency
	}
	if req.IssuedAt != nil {
		updated.IssuedAt = *req.IssuedAt
	}
	// Description is a *string — nil means "not supplied, keep existing";
	// a pointer to any string (including "") means "set to this value".
	if req.Description != nil {
		updated.Description = *req.Description
	}
	if req.RawText != "" {
		updated.RawText = req.RawText
	}

	saved, err := h.invoiceSvc.Update(r.Context(), updated)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// DELETE /api/v1/invoices/{id}
func (h *Handler) deleteInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}
	if err := h.invoiceSvc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, entity.ErrInvoiceNotFound) {
			writeError(w, http.StatusNotFound, "invoice not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/invoices/{id}/download
func (h *Handler) downloadInvoiceFile(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}
	content, mimeType, fileName, err := h.invoiceSvc.GetOriginalFile(r.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrInvoiceNotFound) {
			writeError(w, http.StatusNotFound, "invoice not found")
			return
		}
		writeError(w, http.StatusNotFound, "no file stored for this invoice")
		return
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if fileName == "" {
		fileName = "invoice"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// ── small helper ──────────────────────────────────────────────────────────────

func parseUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}
