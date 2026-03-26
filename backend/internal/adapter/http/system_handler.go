package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
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

	info := map[string]string{
		"storage_mode": h.storageMode,
		"db_host":      h.dbHost,
		"db_state":     dbState,
		"version":      "1.0.0",
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	if h.settingsSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "settings service not available")
		return
	}

	settings, err := h.settingsSvc.GetAll(r.Context())
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

	if err := h.settingsSvc.UpdateMultiple(r.Context(), payload); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update settings")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
