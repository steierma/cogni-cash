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

type DiscoveryHandler struct {
	AppCtx context.Context
	Logger *slog.Logger
	WaitGroup *sync.WaitGroup
	discoverySvc port.DiscoveryUseCase
}

func NewDiscoveryHandler(AppCtx context.Context, Logger *slog.Logger, WaitGroup *sync.WaitGroup, discoverySvc port.DiscoveryUseCase) *DiscoveryHandler {
	return &DiscoveryHandler{
		AppCtx: AppCtx,
		Logger: Logger,
		WaitGroup: WaitGroup,
		discoverySvc: discoverySvc,
	}
}

type ApproveRequest struct {
	Suggestion entity.SuggestedSubscription `json:"suggestion"`
}

func (h *DiscoveryHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	subs, err := h.discoverySvc.ListSubscriptions(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to list subscriptions", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}

	writeJSON(w, http.StatusOK, subs)
}

func (h *DiscoveryHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	sub, err := h.discoverySvc.GetSubscription(r.Context(), subID, userID)
	if err != nil {
		h.Logger.Error("failed to get subscription", "error", err, "user_id", userID, "sub_id", subID)
		writeError(w, http.StatusInternalServerError, "failed to get subscription")
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

func (h *DiscoveryHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	var sub entity.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub.ID = subID
	sub.UserID = userID

	updated, err := h.discoverySvc.UpdateSubscription(r.Context(), sub)
	if err != nil {
		h.Logger.Error("failed to update subscription", "error", err, "user_id", userID, "sub_id", subID)
		writeError(w, http.StatusInternalServerError, "failed to update subscription")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *DiscoveryHandler) GetSuggestedSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	suggestions, err := h.discoverySvc.GetSuggestedSubscriptions(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get suggested subscriptions", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to get suggested subscriptions")
		return
	}

	writeJSON(w, http.StatusOK, suggestions)
}

func (h *DiscoveryHandler) ApproveSubscription(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	subscription, err := h.discoverySvc.ApproveSubscription(r.Context(), userID, req.Suggestion)
	if err != nil {
		h.Logger.Error("failed to approve subscription", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to approve subscription")
		return
	}

	// Trigger AI enrichment asynchronously
	if h.WaitGroup != nil {
		h.WaitGroup.Add(1)
		go func() {
			defer h.WaitGroup.Done()
			h.Logger.Info("Triggering background AI enrichment for approved subscription", "sub_id", subscription.ID, "user_id", userID)
			_, err := h.discoverySvc.EnrichSubscription(h.AppCtx, userID, subscription.ID)
			if err != nil {
				h.Logger.Warn("Background AI enrichment failed", "sub_id", subscription.ID, "error", err)
			} else {
				h.Logger.Info("Background AI enrichment successful", "sub_id", subscription.ID)
			}
		}()
	}

	writeJSON(w, http.StatusCreated, subscription)
}

func (h *DiscoveryHandler) DeclineSubscription(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		MerchantName string `json:"merchant_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MerchantName == "" {
		writeError(w, http.StatusBadRequest, "merchant_name is required")
		return
	}

	err := h.discoverySvc.DeclineSuggestion(r.Context(), userID, req.MerchantName)
	if err != nil {
		h.Logger.Error("failed to decline subscription", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to decline subscription")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "declined"})
}

func (h *DiscoveryHandler) GetDiscoveryFeedback(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	feedback, err := h.discoverySvc.GetDiscoveryFeedback(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get discovery feedback", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to get discovery feedback")
		return
	}

	writeJSON(w, http.StatusOK, feedback)
}

func (h *DiscoveryHandler) RemoveDiscoveryFeedback(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		MerchantName string `json:"merchant_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MerchantName == "" {
		writeError(w, http.StatusBadRequest, "merchant_name is required")
		return
	}

	err := h.discoverySvc.RemoveDiscoveryFeedback(r.Context(), userID, req.MerchantName)
	if err != nil {
		h.Logger.Error("failed to remove discovery feedback", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to remove discovery feedback")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *DiscoveryHandler) EnrichSubscription(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	subscription, err := h.discoverySvc.EnrichSubscription(r.Context(), userID, subID)
	if err != nil {
		h.Logger.Error("failed to enrich subscription", "error", err, "user_id", userID, "sub_id", subID)
		writeError(w, http.StatusInternalServerError, "failed to enrich subscription")
		return
	}

	writeJSON(w, http.StatusOK, subscription)
}

func (h *DiscoveryHandler) PreviewCancellation(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	lang := r.URL.Query().Get("lang")
	draft, err := h.discoverySvc.PreviewCancellation(r.Context(), userID, subID, lang)
	if err != nil {
		h.Logger.Error("failed to preview cancellation", "error", err, "user_id", userID, "sub_id", subID)
		writeError(w, http.StatusInternalServerError, "failed to preview cancellation")
		return
	}

	writeJSON(w, http.StatusOK, draft)
}

func (h *DiscoveryHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	var req struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Subject == "" || req.Body == "" {
		writeError(w, http.StatusBadRequest, "subject and body are required")
		return
	}

	err = h.discoverySvc.CancelSubscription(r.Context(), userID, subID, req.Subject, req.Body)
	if err != nil {
		h.Logger.Error("failed to cancel subscription", "error", err, "user_id", userID, "sub_id", subID)
		writeError(w, http.StatusInternalServerError, "failed to cancel subscription")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "cancellation_sent"})
}

func (h *DiscoveryHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	err = h.discoverySvc.DeleteSubscription(r.Context(), userID, subID)
	if err != nil {
		h.Logger.Error("failed to delete subscription", "error", err, "user_id", userID, "sub_id", subID)
		writeError(w, http.StatusInternalServerError, "failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DiscoveryHandler) GetSubscriptionEvents(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	events, err := h.discoverySvc.GetSubscriptionEvents(r.Context(), userID, subID)
	if err != nil {
		h.Logger.Error("failed to get subscription events", "error", err, "user_id", userID, "sub_id", subID)
		writeError(w, http.StatusInternalServerError, "failed to get subscription events")
		return
	}

	writeJSON(w, http.StatusOK, events)
}

func (h *DiscoveryHandler) CreateSubscriptionFromTransaction(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		TransactionHash string `json:"transaction_hash"`
		MerchantName    string `json:"merchant_name"`
		BillingCycle    string `json:"billing_cycle"`
		BillingInterval int    `json:"billing_interval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TransactionHash == "" || req.BillingCycle == "" {
		writeError(w, http.StatusBadRequest, "transaction_hash and billing_cycle are required")
		return
	}

	if req.BillingInterval <= 0 {
		req.BillingInterval = 1
	}

	sub, err := h.discoverySvc.CreateSubscriptionFromTransaction(r.Context(), userID, req.TransactionHash, req.MerchantName, req.BillingCycle, req.BillingInterval)
	if err != nil {
		h.Logger.Error("failed to create subscription from transaction", "error", err, "user_id", userID, "hash", req.TransactionHash)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

func (h *DiscoveryHandler) LinkTransactions(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	var req struct {
		Hashes []string `json:"hashes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Hashes) == 0 {
		writeError(w, http.StatusBadRequest, "hashes are required")
		return
	}

	err = h.discoverySvc.LinkTransactions(r.Context(), userID, subID, req.Hashes)
	if err != nil {
		h.Logger.Error("failed to link transactions to subscription", "error", err, "user_id", userID, "sub_id", subID)
		writeError(w, http.StatusInternalServerError, "failed to link transactions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "linked"})
}

func (h *DiscoveryHandler) LinkTransaction(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	txnHash := chi.URLParam(r, "hash")
	if txnHash == "" {
		writeError(w, http.StatusBadRequest, "transaction hash is required")
		return
	}

	err = h.discoverySvc.LinkTransaction(r.Context(), userID, subID, txnHash)
	if err != nil {
		h.Logger.Error("failed to link transaction to subscription", "error", err, "user_id", userID, "sub_id", subID, "hash", txnHash)
		writeError(w, http.StatusInternalServerError, "failed to link transaction")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "linked"})
}

func (h *DiscoveryHandler) UnlinkTransaction(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	subIDStr := chi.URLParam(r, "id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	txnHash := chi.URLParam(r, "hash")
	if txnHash == "" {
		writeError(w, http.StatusBadRequest, "transaction hash is required")
		return
	}

	err = h.discoverySvc.UnlinkTransaction(r.Context(), userID, subID, txnHash)
	if err != nil {
		h.Logger.Error("failed to unlink transaction from subscription", "error", err, "user_id", userID, "sub_id", subID, "hash", txnHash)
		writeError(w, http.StatusInternalServerError, "failed to unlink transaction")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "unlinked"})
}
