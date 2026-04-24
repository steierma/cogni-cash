package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (h *Handler) getForecast(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Defaults to 30 days ahead
	now := time.Now()
	fromDate := now
	toDate := now.AddDate(0, 0, 30)

	// Optional query parameters to override dates
	if qFrom := r.URL.Query().Get("from"); qFrom != "" {
		if t, err := time.Parse("2006-01-02", qFrom); err == nil {
			fromDate = t
		}
	}
	if qTo := r.URL.Query().Get("to"); qTo != "" {
		if t, err := time.Parse("2006-01-02", qTo); err == nil {
			toDate = t
		}
	}

	forecast, err := h.forecastingSvc.GetCashFlowForecast(r.Context(), userID, fromDate, toDate)
	if err != nil {
		h.Logger.Error("Failed to generate cash flow forecast", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to generate forecast")
		return
	}

	writeJSON(w, http.StatusOK, forecast)
}
