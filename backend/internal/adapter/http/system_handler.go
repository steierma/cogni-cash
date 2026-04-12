package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

func (h *Handler) healthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getSystemInfo(w http.ResponseWriter, r *http.Request) {
	dbState := "disconnected"
	if h.dbPinger != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.dbPinger(ctx); err == nil {
			dbState = "connected"
		} else {
			dbState = "error: " + err.Error()
		}
	}

	bankProvider := "enablebanking"
	if h.settingsSvc != nil {
		userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
		if val, err := h.settingsSvc.Get(r.Context(), "bank_provider", userID); err == nil && val != "" {
			bankProvider = val
		}
	}
	if bankProvider == "" {
		bankProvider = os.Getenv("BANK_PROVIDER")
		if bankProvider == "" {
			bankProvider = "enablebanking"
		}
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "unknown"
	}

	info := map[string]string{
		"storage_mode":  h.storageMode,
		"db_host":       h.dbHost,
		"db_state":      dbState,
		"version":       version,
		"bank_provider": bankProvider,
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	if h.settingsSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "settings service not available")
		return
	}

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	settings, err := h.settingsSvc.GetAll(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	if h.settingsSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "settings service not available")
		return
	}

	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.settingsSvc.UpdateMultiple(r.Context(), payload, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update settings")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) sendTestEmail(w http.ResponseWriter, r *http.Request) {
	if h.notificationSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "notification service not available")
		return
	}

	var payload struct {
		To string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if payload.To == "" {
		writeError(w, http.StatusBadRequest, "recipient email is required")
		return
	}

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.notificationSvc.SendTestEmail(r.Context(), payload.To, userID); err != nil {
		h.Logger.Error("Test email failed", "to", payload.To, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to send test email: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Test email sent successfully"})
}
