package http

import (
	"cogni-cash/internal/domain/entity"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"log/slog"
	"cogni-cash/internal/domain/port"
)

type DocumentHandler struct {
	Logger *slog.Logger
	documentSvc port.DocumentUseCase
}

func NewDocumentHandler(Logger *slog.Logger, documentSvc port.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{
		Logger: Logger,
		documentSvc: documentSvc,
	}
}

// uploadDocument handles POST /api/v1/documents/upload/
func (h *DocumentHandler) uploadDocument(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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
func (h *DocumentHandler) listDocuments(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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
func (h *DocumentHandler) getDocument(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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
func (h *DocumentHandler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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
func (h *DocumentHandler) downloadDocument(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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

	// Diagnostic: Check if we are sending PGP data as PDF (means decryption failed)
	if len(content) > 3 && content[0] == 0xc3 && content[1] == 0x0d && content[2] == 0x04 {
		h.Logger.Error("CRITICAL: Sending encrypted PGP data as raw document. Decryption definitely failed. Please check DOCUMENT_VAULT_KEY.", "id", id)
	}

	h.Logger.Info("Sending document to browser", "id", id, "mime_type", mimeType, "size", len(content), "filename", fileName)

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// getTaxYearSummary handles GET /api/v1/documents/tax-summary/{year}/
func (h *DocumentHandler) getTaxYearSummary(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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

func (h *DocumentHandler) updateDocument(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	updateReq := entity.DocumentUpdateRequest{}

	// Check content type for multipart (file upload) or JSON
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		const maxUpload = 20 << 20
		r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
		if err := r.ParseMultipartForm(maxUpload); err != nil {
			writeError(w, http.StatusRequestEntityTooLarge, "upload too large")
			return
		}

		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			fileBytes, err := io.ReadAll(file)
			if err == nil {
				updateReq.OriginalFileContent = fileBytes
				mimeType := resolveMIME(header.Header.Get("Content-Type"), header.Filename)
				updateReq.MimeType = &mimeType
			}
		}

		if val := r.FormValue("file_name"); val != "" {
			updateReq.OriginalFileName = &val
		}
		if val := r.FormValue("document_type"); val != "" {
			dt := entity.DocumentType(val)
			updateReq.Type = &dt
		}
		if val := r.FormValue("document_date"); val != "" {
			if parsedDate, err := time.Parse("2006-01-02", val); err == nil {
				updateReq.DocumentDate = &parsedDate
			}
		}
	} else {
		var req struct {
			FileName     *string                `json:"file_name"`
			Type         *entity.DocumentType   `json:"type"`
			DocumentDate *string                `json:"document_date"`
			Metadata     map[string]interface{} `json:"metadata"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "Invalid request body")
			return
		}

		updateReq.OriginalFileName = req.FileName
		updateReq.Type = req.Type
		updateReq.Metadata = req.Metadata

		if req.DocumentDate != nil && *req.DocumentDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", *req.DocumentDate); err == nil {
				updateReq.DocumentDate = &parsedDate
			}
		}
	}

	doc, err := h.documentSvc.Update(r.Context(), id, userID, updateReq)
	if err != nil {
		h.Logger.Error("Failed to update document", "id", id, "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to update document")
		return
	}

	writeJSON(w, http.StatusOK, doc)
}
