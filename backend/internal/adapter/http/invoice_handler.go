package http

import (
	"encoding/json"
	"net/http"

	"cogni-cash/internal/domain/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type categorizeDocumentRequest struct {
	RawText string `json:"raw_text"`
}

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.invoiceRepo.FindAll(r.Context())
	if err != nil {
		h.Logger.Error("Failed to list invoices", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, invoices)
}

func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	invoice, err := h.invoiceRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, invoice)
}

func (h *Handler) categorizeDocument(w http.ResponseWriter, r *http.Request) {
	var req categorizeDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	invoice, err := h.categorizationSvc.CategorizeDocument(r.Context(), req.RawText)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrEmptyRawText {
			status = http.StatusUnprocessableEntity
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, invoice)
}

func (h *Handler) deleteInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	if err := h.invoiceRepo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
