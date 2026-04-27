package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"log/slog"
)

type InvoiceHandler struct {
	Logger *slog.Logger
	invoiceSvc port.InvoiceUseCase
}

func NewInvoiceHandler(Logger *slog.Logger, invoiceSvc port.InvoiceUseCase) *InvoiceHandler {
	return &InvoiceHandler{
		Logger: Logger,
		invoiceSvc: invoiceSvc,
	}
}

// allowedMIMETypes is the set of MIME types accepted by the invoice import endpoint.
var allowedMIMETypes = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
}

// ── request / response types ─────────────────────────────────────────────────

type updateInvoiceRequest struct {
	VendorName  string                `json:"vendor_name"`
	CategoryID  *uuid.UUID            `json:"category_id"`
	Amount      float64               `json:"amount"`
	Currency    string                `json:"currency"`
	IssuedAt    *time.Time            `json:"issued_at"`
	Description *string               `json:"description"` // pointer so empty-string clears the field
	Splits      []entity.InvoiceSplit `json:"splits"`
}

type importManualRequest struct {
	VendorName  string                `json:"vendor_name"`
	CategoryID  *uuid.UUID            `json:"category_id"`
	Amount      float64               `json:"amount"`
	Currency    string                `json:"currency"`
	IssuedAt    time.Time             `json:"issued_at"`
	Description string                `json:"description"`
	Splits      []entity.InvoiceSplit `json:"splits"`
}

type shareInvoiceRequest struct {
	UserID     string `json:"user_id"`
	Permission string `json:"permission"` // "view" or "edit"
}

type updateInvoicesBulkRequest struct {
	IDs        []uuid.UUID `json:"ids"`
	CategoryID *uuid.UUID  `json:"category_id"`
}

// ── handlers ──────────────────────────────────────────────────────────────────

// POST /api/v1/invoices/{id}/share/
func (h *InvoiceHandler) shareInvoice(w http.ResponseWriter, r *http.Request) {
	if h.invoiceSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "invoice service not available")
		return
	}

	invoiceID, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	var req shareInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sharedWithID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	ownerID := GetUserID(r.Context())
	if ownerID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	perm := req.Permission
	if perm == "" {
		perm = "view"
	}

	if err := h.invoiceSvc.ShareInvoice(r.Context(), invoiceID, ownerID, sharedWithID, perm); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/invoices/{id}/share/{user_id}/
func (h *InvoiceHandler) revokeInvoiceShare(w http.ResponseWriter, r *http.Request) {
	if h.invoiceSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "invoice service not available")
		return
	}

	invoiceID, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	sharedWithID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.invoiceSvc.RevokeInvoiceShare(r.Context(), invoiceID, userID, sharedWithID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/invoices/{id}/shares/
func (h *InvoiceHandler) listInvoiceShares(w http.ResponseWriter, r *http.Request) {
	if h.invoiceSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "invoice service not available")
		return
	}

	invoiceID, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	ownerID := GetUserID(r.Context())
	if ownerID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	shares, err := h.invoiceSvc.ListInvoiceShares(r.Context(), invoiceID, ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, shares)
}

// PATCH /api/v1/invoices/bulk-category/
func (h *InvoiceHandler) updateInvoicesBulk(w http.ResponseWriter, r *http.Request) {
	if h.invoiceSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "invoice service not available")
		return
	}

	var req updateInvoicesBulkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "missing invoice ids")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.invoiceSvc.UpdateCategoriesBulk(r.Context(), req.IDs, req.CategoryID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/invoices/
func (h *InvoiceHandler) listInvoices(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter := entity.InvoiceFilter{
		UserID: userID,
	}
	q := r.URL.Query()
	if q.Get("include_shared") == "true" {
		filter.IncludeShared = true
	}
	if src := q.Get("source"); src != "" {
		filter.Source = src
	}
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = offset
	}

	invoices, err := h.invoiceSvc.GetAll(r.Context(), filter)
	if err != nil {
		h.Logger.Error("Failed to list invoices", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if invoices == nil {
		invoices = []entity.Invoice{}
	}
	writeJSON(w, http.StatusOK, invoices)
}

// GET /api/v1/invoices/{id}
func (h *InvoiceHandler) getInvoice(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}
	invoice, err := h.invoiceSvc.GetByID(r.Context(), id, userID)
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
// Accepts: application/pdf, image/jpeg, image/png, image/gif, image/webp
func (h *InvoiceHandler) importInvoice(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

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

	// Resolve the MIME type via the shared helper (prefers Content-Type header,
	// falls back to extension-based detection).
	mimeType := resolveMIME(header.Header.Get("Content-Type"), header.Filename)

	// Default to PDF if we still could not determine the type
	if mimeType == "application/octet-stream" {
		mimeType = "application/pdf"
	}

	// Validate against the allowed set
	if !allowedMIMETypes[mimeType] {
		writeError(w, http.StatusUnsupportedMediaType,
			fmt.Sprintf("unsupported file type %q — accepted types: PDF, JPEG, PNG, GIF, WEBP", mimeType))
		return
	}

	// ── Manual Overrides ──
	var overrides port.ImportOverrides

	if v := r.FormValue("vendor_name"); v != "" {
		overrides.VendorName = &v
	}
	if v := r.FormValue("amount"); v != "" {
		if amt, err := strconv.ParseFloat(v, 64); err == nil {
			overrides.Amount = &amt
		}
	}
	if v := r.FormValue("currency"); v != "" {
		overrides.Currency = &v
	}
	if v := r.FormValue("issued_at"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			overrides.IssuedAt = &t
		}
	}
	if v := r.FormValue("category_id"); v != "" && v != "null" {
		if id, err := uuid.Parse(v); err == nil {
			overrides.CategoryID = &id
		}
	}
	if v := r.FormValue("splits"); v != "" {
		if err := json.Unmarshal([]byte(v), &overrides.Splits); err != nil {
			writeError(w, http.StatusBadRequest, "invalid splits format (JSON expected)")
			return
		}
	}

	invoice, err := h.invoiceSvc.ImportFromFile(r.Context(), userID, header.Filename, mimeType, fileBytes, overrides)
	if err != nil {
		if errors.Is(err, entity.ErrInvoiceDuplicate) {
			writeError(w, http.StatusConflict, "invoice already imported (duplicate)")
			return
		}
		h.Logger.Error("Failed to import invoice", "error", err, "user_id", userID)
		writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("import failed: %s", err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, invoice)
}

// POST /api/v1/invoices/ (manual import without file)
func (h *InvoiceHandler) importManual(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req importManualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Warn("importManual failed: invalid JSON body", "user_id", userID, "error", err)
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err.Error()))
		return
	}

	invoice := entity.Invoice{
		ID:          uuid.New(),
		UserID:      userID,
		Vendor:      entity.Vendor{ID: uuid.New(), Name: req.VendorName},
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		IssuedAt:    req.IssuedAt,
		Description: req.Description,
		Splits:      req.Splits,
	}

	saved, err := h.invoiceSvc.ImportManual(r.Context(), userID, invoice)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, saved)
}

// PUT /api/v1/invoices/{id}
func (h *InvoiceHandler) updateInvoice(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	var req updateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Warn("updateInvoice failed: invalid JSON body", "invoice_id", id, "error", err)
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err.Error()))
		return
	}

	// Fetch current record to preserve immutable fields and verify ownership
	existing, err := h.invoiceSvc.GetByID(r.Context(), id, userID)
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

	// Always overwrite splits if they are supplied in the request.
	// We check for nil so that clients who don't support splits yet don't wipe them.
	if req.Splits != nil {
		updated.Splits = req.Splits
	}

	saved, err := h.invoiceSvc.Update(r.Context(), updated)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// DELETE /api/v1/invoices/{id}
func (h *InvoiceHandler) deleteInvoice(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}
	if err := h.invoiceSvc.Delete(r.Context(), id, userID); err != nil {
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
func (h *InvoiceHandler) downloadInvoiceFile(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}
	content, mimeType, fileName, err := h.invoiceSvc.GetOriginalFile(r.Context(), id, userID)
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
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, fileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// ── small helper ──────────────────────────────────────────────────────────────

func parseUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}
