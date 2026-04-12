package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createBridgeTokenRequest struct {
	Name string `json:"name"`
}

func (h *Handler) listBridgeTokens(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tokens, err := h.bridgeTokenSvc.ListTokens(r.Context(), userID)
	if err != nil {
		h.Logger.Error("Failed to list bridge tokens", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to list bridge tokens")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (h *Handler) createBridgeToken(w http.ResponseWriter, r *http.Request) {
	var req createBridgeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	response, err := h.bridgeTokenSvc.CreateToken(r.Context(), userID, req.Name)
	if err != nil {
		h.Logger.Error("Failed to create bridge token", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to create bridge token")
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) revokeBridgeToken(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid token id")
		return
	}

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.bridgeTokenSvc.RevokeToken(r.Context(), id, userID); err != nil {
		h.Logger.Error("Failed to revoke bridge token", "error", err, "user_id", userID, "token_id", id)
		writeError(w, http.StatusInternalServerError, "failed to revoke bridge token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
