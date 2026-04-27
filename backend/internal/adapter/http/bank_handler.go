package http

import (
	"cogni-cash/internal/domain/entity"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"context"
	"sync"
	"log/slog"
	"cogni-cash/internal/domain/port"
)

type BankHandler struct {
	AppCtx context.Context
	Logger *slog.Logger
	WaitGroup *sync.WaitGroup
	bankSvc port.BankUseCase
}

func NewBankHandler(AppCtx context.Context, Logger *slog.Logger, WaitGroup *sync.WaitGroup, bankSvc port.BankUseCase) *BankHandler {
	return &BankHandler{
		AppCtx: AppCtx,
		Logger: Logger,
		WaitGroup: WaitGroup,
		bankSvc: bankSvc,
	}
}

func (h *BankHandler) listBankInstitutions(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	if country == "" {
		country = "DE"
	}

	isSandbox := r.URL.Query().Get("sandbox") == "true"

	userID := GetUserID(r.Context())
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

func (h *BankHandler) createBankConnection(w http.ResponseWriter, r *http.Request) {
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

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ip := getClientIP(r)
	ua := r.Header.Get("User-Agent")

	conn, err := h.bankSvc.CreateConnection(r.Context(), userID, req.InstitutionID, req.InstitutionName, req.Country, req.RedirectURL, req.Sandbox, ip, ua)
	if err != nil {
		h.Logger.Error("failed to create bank connection", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to initiate bank link")
		return
	}

	writeJSON(w, http.StatusCreated, conn)
}

func (h *BankHandler) finishBankConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RequisitionID string `json:"requisition_id"`
		Code          string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := GetUserID(r.Context())
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

func (h *BankHandler) createVirtualBankAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		IBAN        string  `json:"iban"`
		Currency    string  `json:"currency"`
		AccountType string  `json:"account_type"`
		Balance     float64 `json:"balance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	acc := &entity.BankAccount{
		UserID:      userID,
		Name:        req.Name,
		IBAN:        req.IBAN,
		Currency:    req.Currency,
		AccountType: entity.StatementType(req.AccountType),
		Balance:     req.Balance,
	}

	if err := h.bankSvc.CreateVirtualAccount(r.Context(), acc); err != nil {
		h.Logger.Error("failed to create virtual bank account", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create virtual account")
		return
	}

	writeJSON(w, http.StatusCreated, acc)
}

func (h *BankHandler) listBankConnections(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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

func (h *BankHandler) deleteBankConnection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid connection id")
		return
	}

	userID := GetUserID(r.Context())
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

func (h *BankHandler) refreshBankConnection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid connection id")
		return
	}

	var req struct {
		RedirectURL string `json:"redirect_url"`
		Sandbox     bool   `json:"sandbox"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ip := getClientIP(r)
	ua := r.Header.Get("User-Agent")

	conn, err := h.bankSvc.RefreshConnection(r.Context(), id, userID, req.RedirectURL, req.Sandbox, ip, ua)
	if err != nil {
		h.Logger.Error("failed to refresh bank connection", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to initiate bank refresh")
		return
	}

	writeJSON(w, http.StatusOK, conn)
}

func (h *BankHandler) syncAllBankAccounts(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
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

func (h *BankHandler) updateBankAccountType(w http.ResponseWriter, r *http.Request) {
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

	userID := GetUserID(r.Context())
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

func (h *BankHandler) shareBankAccount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	var req struct {
		SharedWithID uuid.UUID `json:"shared_with_user_id"`
		Permission   string    `json:"permission_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := GetUserID(r.Context())
	if err := h.bankSvc.ShareAccount(r.Context(), accountID, userID, req.SharedWithID, req.Permission); err != nil {
		h.Logger.Error("Failed to share bank account", "error", err, "account_id", accountID)
		writeError(w, http.StatusInternalServerError, "failed to share bank account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "shared"})
}

func (h *BankHandler) revokeBankAccountShare(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	sharedWithStr := chi.URLParam(r, "user_id")
	sharedWithID, err := uuid.Parse(sharedWithStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	userID := GetUserID(r.Context())
	if err := h.bankSvc.RevokeShare(r.Context(), accountID, userID, sharedWithID); err != nil {
		h.Logger.Error("Failed to revoke bank account share", "error", err, "account_id", accountID)
		writeError(w, http.StatusInternalServerError, "failed to revoke share")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BankHandler) listBankAccountShares(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	userID := GetUserID(r.Context())
	shares, err := h.bankSvc.ListShares(r.Context(), accountID, userID)
	if err != nil {
		h.Logger.Error("Failed to list bank account shares", "error", err, "account_id", accountID)
		writeError(w, http.StatusInternalServerError, "failed to list shares")
		return
	}

	writeJSON(w, http.StatusOK, shares)
}
