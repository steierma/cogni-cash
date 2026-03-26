package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"cogni-cash/internal/domain/entity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createReconciliationRequest struct {
	SettlementTxHash string `json:"settlement_tx_hash"`
	TargetTxHash     string `json:"target_tx_hash"`
}

func (h *Handler) getReconciliationSuggestions(w http.ResponseWriter, r *http.Request) {
	if h.reconciliationSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "reconciliation service not available")
		return
	}

	matchWindowDays := 7
	windowQuery := r.URL.Query().Get("window")
	if windowQuery != "" {
		if parsed, err := strconv.Atoi(windowQuery); err == nil && parsed > 0 {
			matchWindowDays = parsed
		}
	}

	suggestions, err := h.reconciliationSvc.SuggestReconciliations(r.Context(), matchWindowDays)
	if err != nil {
		h.Logger.Error("Failed to get reconciliation suggestions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get suggestions")
		return
	}

	writeJSON(w, http.StatusOK, suggestions)
}

func (h *Handler) createReconciliation(w http.ResponseWriter, r *http.Request) {
	if h.reconciliationSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "reconciliation service not available")
		return
	}

	var req createReconciliationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SettlementTxHash == "" || req.TargetTxHash == "" {
		writeError(w, http.StatusBadRequest, "hashes are required")
		return
	}

	rec, err := h.reconciliationSvc.ReconcileStatements(r.Context(), req.SettlementTxHash, req.TargetTxHash)
	if err != nil {
		if errors.Is(err, entity.ErrTransactionNotFound) {
			writeError(w, http.StatusNotFound, "transaction not found")
			return
		}
		if errors.Is(err, entity.ErrSameAccount) {
			writeError(w, http.StatusBadRequest, "source and target must be from different accounts")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create reconciliation")
		return
	}

	writeJSON(w, http.StatusCreated, rec)
}

func (h *Handler) listReconciliations(w http.ResponseWriter, r *http.Request) {
	if h.reconciliationRepo == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	recs, err := h.reconciliationRepo.FindAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list reconciliations")
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (h *Handler) deleteReconciliation(w http.ResponseWriter, r *http.Request) {
	if h.reconciliationSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "reconciliation service not available")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid reconciliation ID")
		return
	}

	if err := h.reconciliationSvc.DeleteReconciliation(r.Context(), id); err != nil {
		h.Logger.Error("Failed to delete reconciliation", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete reconciliation")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
