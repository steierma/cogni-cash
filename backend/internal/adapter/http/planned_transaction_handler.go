package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"cogni-cash/internal/domain/entity"
)

type createPlannedTransactionRequest struct {
	Amount             float64                   `json:"amount"`
	Currency           string                    `json:"currency"`
	Date               time.Time                 `json:"date"`
	Description        string                    `json:"description"`
	CategoryID         *uuid.UUID                `json:"category_id"`
	BankAccountID      *uuid.UUID                `json:"bank_account_id"`
	IntervalMonths     int                       `json:"interval_months"`
	SchedulingStrategy entity.SchedulingStrategy `json:"scheduling_strategy"`
	EndDate            *time.Time                `json:"end_date"`
}

type updatePlannedTransactionRequest struct {
	Amount             float64                         `json:"amount"`
	Currency           string                          `json:"currency"`
	Date               time.Time                       `json:"date"`
	Description        string                          `json:"description"`
	CategoryID         *uuid.UUID                      `json:"category_id"`
	BankAccountID      *uuid.UUID                      `json:"bank_account_id"`
	Status             entity.PlannedTransactionStatus `json:"status"`
	IntervalMonths     int                             `json:"interval_months"`
	SchedulingStrategy entity.SchedulingStrategy       `json:"scheduling_strategy"`
	EndDate            *time.Time                      `json:"end_date"`
}

func (h *Handler) listPlannedTransactions(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	pts, err := h.plannedTransactionSvc.FindByUserID(r.Context(), userID)
	if err != nil {
		h.Logger.Error("Failed to list planned transactions", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if pts == nil {
		pts = []entity.PlannedTransaction{}
	}
	writeJSON(w, http.StatusOK, pts)
}

func (h *Handler) createPlannedTransaction(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createPlannedTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	pt := &entity.PlannedTransaction{
		UserID:             userID,
		Amount:             req.Amount,
		Currency:           req.Currency,
		Date:               req.Date,
		Description:        req.Description,
		CategoryID:         req.CategoryID,
		BankAccountID:      req.BankAccountID,
		IntervalMonths:     req.IntervalMonths,
		SchedulingStrategy: req.SchedulingStrategy,
		EndDate:            req.EndDate,
	}

	if err := h.plannedTransactionSvc.Create(r.Context(), pt); err != nil {
		if err == entity.ErrInvalidPlannedTransaction {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.Logger.Error("Failed to create planned transaction", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to create planned transaction")
		return
	}

	writeJSON(w, http.StatusCreated, pt)
}

func (h *Handler) updatePlannedTransaction(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	var req updatePlannedTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	pt := &entity.PlannedTransaction{
		ID:                 id,
		UserID:             userID,
		Amount:             req.Amount,
		Currency:           req.Currency,
		Date:               req.Date,
		Description:        req.Description,
		CategoryID:         req.CategoryID,
		BankAccountID:      req.BankAccountID,
		Status:             req.Status,
		IntervalMonths:     req.IntervalMonths,
		SchedulingStrategy: req.SchedulingStrategy,
		EndDate:            req.EndDate,
	}

	if err := h.plannedTransactionSvc.Update(r.Context(), pt); err != nil {
		if err == entity.ErrPlannedTransactionNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == entity.ErrInvalidPlannedTransaction {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.Logger.Error("Failed to update planned transaction", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to update planned transaction")
		return
	}

	writeJSON(w, http.StatusOK, pt)
}

func (h *Handler) deletePlannedTransaction(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	if err := h.plannedTransactionSvc.Delete(r.Context(), id, userID); err != nil {
		h.Logger.Error("Failed to delete planned transaction", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "Failed to delete planned transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
