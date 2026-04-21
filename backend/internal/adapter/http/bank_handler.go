package http

import (
	"cogni-cash/internal/domain/entity"
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

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	insts, err := h.bankSvc.GetInstitutions(r.Context(), userID, country, isSandbox)
	if err != nil {
		h.Logger.Error("failed to list institutions", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch bank list")
		return
	}

	writeJSON(w, http.StatusOK, insts)
}

func (h *Handler) createBankConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InstitutionID   string `json:"institution_id"`
		InstitutionName string `json:"institution_name"`
		Country         string `json:"country"`
		RedirectURL     string `json:"redirect_url"`
		Sandbox         bool   `json:"sandbox"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ip := h.getClientIP(r)
	ua := r.Header.Get("User-Agent")

	conn, err := h.bankSvc.CreateConnection(r.Context(), userID, req.InstitutionID, req.InstitutionName, req.Country, req.RedirectURL, req.Sandbox, ip, ua)
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

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.bankSvc.FinishConnection(r.Context(), userID, req.RequisitionID, req.Code); err != nil {
		h.Logger.Error("failed to finish bank connection", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to complete bank link")
		return
	}

	// Trigger background sync so user sees data immediately after connecting
	if h.WaitGroup != nil {
		h.WaitGroup.Add(1)
		go func() {
			defer h.WaitGroup.Done()
			if err := h.bankSvc.SyncAllAccounts(h.AppCtx, userID); err != nil {
				h.Logger.Error("background sync failed after connection finish", "user_id", userID, "error", err)
			}
		}()
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) listBankConnections(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
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

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.bankSvc.DeleteConnection(r.Context(), id, userID); err != nil {
		h.Logger.Error("failed to delete bank connection", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete connection")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) syncAllBankAccounts(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if h.WaitGroup != nil {
		h.WaitGroup.Add(1)
		go func() {
			defer h.WaitGroup.Done()
			if err := h.bankSvc.SyncAllAccounts(h.AppCtx, userID); err != nil {
				h.Logger.Error("background sync failed", "user_id", userID, "error", err)
			}
		}()
	}

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

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.bankSvc.UpdateAccountType(r.Context(), id, entity.StatementType(req.AccountType), userID); err != nil {
		h.Logger.Error("failed to update account type", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update account type")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
