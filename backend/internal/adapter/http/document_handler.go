package http

import (
	"cogni-cash/internal/domain/entity"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// uploadDocument handles POST /api/v1/documents/upload/
func (h *Handler) uploadDocument(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	const maxUpload = 20 << 20 // 20 MB hard cap for vault
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "upload too large (max 20 MB)")
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

	skipAI := r.FormValue("skip_ai") == "true"

	// Default to other if manual type is empty
	docTypeForm := r.FormValue("document_type")
	if docTypeForm == "" {
		docTypeForm = string(entity.DocTypeOther)
	}
	docType := entity.DocumentType(docTypeForm)

	docDateStr := r.FormValue("document_date")
	fileName := r.FormValue("file_name")
	if fileName == "" {
		fileName = header.Filename
	}

	var docDate time.Time
	if docDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", docDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid date format (expected YYYY-MM-DD)")
			return
		}
		docDate = parsedDate
	}

	req := entity.DocumentUploadRequest{
		UserID:       userID,
		FileContent:  fileBytes,
		FileName:     fileName,
		ContentType:  mimeType,
		SkipAI:       skipAI,
		ManualType:   docType,
		DocumentDate: docDate,
	}

	doc, err := h.documentSvc.Upload(r.Context(), req)
	if err != nil {
		if errors.Is(err, entity.ErrDocumentDuplicate) {
			writeError(w, http.StatusConflict, "Document already exists (duplicate hash)")
			return
		}
		h.Logger.Error("Failed to upload document", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to upload document")
		return
	}

	writeJSON(w, http.StatusCreated, doc)
}

// listDocuments handles GET /api/v1/documents/
func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter := entity.DocumentFilter{
		UserID: userID,
		Type:   entity.DocumentType(r.URL.Query().Get("type")),
		Search: r.URL.Query().Get("search"),
	}

	docs, err := h.documentSvc.List(r.Context(), filter)
	if err != nil {
		h.Logger.Error("Failed to list documents", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to fetch documents")
		return
	}

	if docs == nil {
		docs = []entity.Document{}
	}
	writeJSON(w, http.StatusOK, docs)
}

// getDocument handles GET /api/v1/documents/{id}/
func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	doc, err := h.documentSvc.GetDetail(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, entity.ErrDocumentNotFound) {
			writeError(w, http.StatusNotFound, "Document not found")
			return
		}
		h.Logger.Error("Failed to fetch document", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to fetch document")
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

// deleteDocument handles DELETE /api/v1/documents/{id}/
func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	err = h.documentSvc.Delete(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, entity.ErrDocumentNotFound) {
			writeError(w, http.StatusNotFound, "Document not found")
			return
		}
		h.Logger.Error("Failed to delete document", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to delete document")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// downloadDocument handles GET /api/v1/documents/{id}/download/
func (h *Handler) downloadDocument(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	content, mimeType, fileName, err := h.documentSvc.Download(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, entity.ErrDocumentNotFound) {
			writeError(w, http.StatusNotFound, "Document not found")
			return
		}
		h.Logger.Error("Failed to download document", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to download document")
		return
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// getTaxYearSummary handles GET /api/v1/documents/tax-summary/{year}/
func (h *Handler) getTaxYearSummary(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var year int
	_, err := fmt.Sscanf(chi.URLParam(r, "year"), "%d", &year)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid year format")
		return
	}

	summary, err := h.documentSvc.GetTaxYearSummary(r.Context(), userID, year)
	if err != nil {
		h.Logger.Error("Failed to fetch tax summary", "year", year, "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to fetch tax summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) updateDocument(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	var req struct {
		FileName     *string                `json:"file_name"`
		Type         *entity.DocumentType   `json:"type"`
		DocumentDate *string                `json:"document_date"`
		Metadata     map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var docDate *time.Time
	if req.DocumentDate != nil && *req.DocumentDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *req.DocumentDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid date format (expected YYYY-MM-DD)")
			return
		}
		docDate = &parsedDate
	}

	updateReq := entity.DocumentUpdateRequest{
		OriginalFileName: req.FileName,
		Type:             req.Type,
		DocumentDate:     docDate,
		Metadata:         req.Metadata,
	}

	doc, err := h.documentSvc.Update(r.Context(), id, userID, updateReq)
	if err != nil {
		h.Logger.Error("Failed to update document", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to update document")
		return
	}

	writeJSON(w, http.StatusOK, doc)
}
