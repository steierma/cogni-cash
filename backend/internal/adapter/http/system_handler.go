package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
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

	// User laden, um Admin-Status für das Filtern der Settings zu prüfen
	user, err := h.userSvc.GetUser(r.Context(), userID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to verify permissions")
		return
	}
	isAdmin := user.Role == "admin"

	settings, err := h.settingsSvc.GetAllMasked(r.Context(), userID, isAdmin)
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

	// Fetch user to check role
	user, err := h.userSvc.GetUser(r.Context(), userID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to verify permissions")
		return
	}

	isAdmin := user.Role == "admin"

	if err := h.settingsSvc.UpdateMultiple(r.Context(), payload, userID, isAdmin); err != nil {
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

func (h *Handler) getLogLevel(w http.ResponseWriter, _ *http.Request) {
	if h.LogLevel == nil {
		writeError(w, http.StatusServiceUnavailable, "log level control not available")
		return
	}

	level := strings.ToUpper(h.LogLevel.Level().String())
	writeJSON(w, http.StatusOK, map[string]string{"level": level})
}

func (h *Handler) updateLogLevel(w http.ResponseWriter, r *http.Request) {
	if h.LogLevel == nil {
		writeError(w, http.StatusServiceUnavailable, "log level control not available")
		return
	}

	var payload struct {
		Level string `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var newLevel slog.Level
	switch strings.ToUpper(payload.Level) {
	case "DEBUG":
		newLevel = slog.LevelDebug
	case "INFO":
		newLevel = slog.LevelInfo
	case "WARN":
		newLevel = slog.LevelWarn
	case "ERROR":
		newLevel = slog.LevelError
	default:
		writeError(w, http.StatusBadRequest, "invalid log level: "+payload.Level)
		return
	}

	h.LogLevel.Set(newLevel)
	h.Logger.Info("Log level changed dynamically", "new_level", newLevel.String())

	// Persistence: Update in settings if settingsSvc is available
	if h.settingsSvc != nil {
		userID := h.getUserID(r.Context())
		if userID != uuid.Nil {
			// Find admin ID to save it globally (for all users)
			adminID, err := h.userSvc.GetAdminID(r.Context())
			if err == nil {
				_ = h.settingsSvc.UpdateMultiple(r.Context(), map[string]string{"log_level": payload.Level}, adminID, true)
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"level": newLevel.String()})
}
