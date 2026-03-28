package http

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) listBankInstitutions(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	if country == "" {
		country = "DE"
	}

	isSandbox := r.URL.Query().Get("sandbox") == "true"

	insts, err := h.bankSvc.GetInstitutions(r.Context(), country, isSandbox)
	if err != nil {
		h.Logger.Error("failed to list institutions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch bank list")
		return
	}

	writeJSON(w, http.StatusOK, insts)
}

func (h *Handler) createBankConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InstitutionID string `json:"institution_id"`
		Country       string `json:"country"`
		RedirectURL   string `json:"redirect_url"`
		Sandbox       bool   `json:"sandbox"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := h.getUserID(r.Context())
	conn, err := h.bankSvc.CreateConnection(r.Context(), userID, req.InstitutionID, req.Country, req.RedirectURL, req.Sandbox)
	if err != nil {
		h.Logger.Error("failed to create bank connection", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to initiate bank link")
		return
	}

	writeJSON(w, http.StatusCreated, conn)
}

func (h *Handler) finishBankConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RequisitionID string `json:"requisition_id"`
		Code          string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.bankSvc.FinishConnection(r.Context(), req.RequisitionID, req.Code); err != nil {
		h.Logger.Error("failed to finish bank connection", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to complete bank link")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) listBankConnections(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	conns, err := h.bankSvc.GetConnections(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to list bank connections", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch connections")
		return
	}

	writeJSON(w, http.StatusOK, conns)
}

func (h *Handler) deleteBankConnection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid connection id")
		return
	}

	if err := h.bankSvc.DeleteConnection(r.Context(), id); err != nil {
		h.Logger.Error("failed to delete bank connection", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete connection")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) syncAllBankAccounts(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())

	go func() {
		ctx := context.Background() // Fresh context for background task
		if err := h.bankSvc.SyncAllAccounts(ctx, userID); err != nil {
			h.Logger.Error("background sync failed", "user_id", userID, "error", err)
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{"message": "sync started in background"})
}

func (h *Handler) updateBankAccountType(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	var req struct {
		AccountType string `json:"account_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.bankSvc.UpdateAccountType(r.Context(), id, entity.StatementType(req.AccountType)); err != nil {
		h.Logger.Error("failed to update account type", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update account type")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value(userIDKey)
	if val == nil {
		return uuid.Nil
	}

	if id, ok := val.(uuid.UUID); ok {
		return id
	}

	if idStr, ok := val.(string); ok {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return uuid.Nil
		}
		return id
	}

	return uuid.Nil
}
