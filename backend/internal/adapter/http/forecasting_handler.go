package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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

func (h *Handler) excludeForecast(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	forecastIDStr := chi.URLParam(r, "id")
	forecastID, err := uuid.Parse(forecastIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid forecast id")
		return
	}

	if err := h.forecastingSvc.ExcludeForecast(r.Context(), userID, forecastID); err != nil {
		h.Logger.Error("Failed to exclude forecast", "error", err, "user_id", userID, "forecast_id", forecastID)
		writeError(w, http.StatusInternalServerError, "failed to exclude forecast")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "excluded"})
}

func (h *Handler) includeForecast(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	forecastIDStr := chi.URLParam(r, "id")
	forecastID, err := uuid.Parse(forecastIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid forecast id")
		return
	}

	if err := h.forecastingSvc.IncludeForecast(r.Context(), userID, forecastID); err != nil {
		h.Logger.Error("Failed to include forecast", "error", err, "user_id", userID, "forecast_id", forecastID)
		writeError(w, http.StatusInternalServerError, "failed to include forecast")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "included"})
}

func (h *Handler) excludePattern(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		MatchTerm string `json:"match_term"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MatchTerm == "" {
		writeError(w, http.StatusBadRequest, "match_term is required")
		return
	}

	if err := h.forecastingSvc.ExcludePattern(r.Context(), userID, req.MatchTerm); err != nil {
		h.Logger.Error("Failed to exclude pattern", "error", err, "user_id", userID, "match_term", req.MatchTerm)
		writeError(w, http.StatusInternalServerError, "failed to exclude pattern")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "excluded"})
}

func (h *Handler) includePattern(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		MatchTerm string `json:"match_term"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MatchTerm == "" {
		writeError(w, http.StatusBadRequest, "match_term is required")
		return
	}

	if err := h.forecastingSvc.IncludePattern(r.Context(), userID, req.MatchTerm); err != nil {
		h.Logger.Error("Failed to include pattern", "error", err, "user_id", userID, "match_term", req.MatchTerm)
		writeError(w, http.StatusInternalServerError, "failed to include pattern")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "included"})
}

func (h *Handler) listPatternExclusions(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	exclusions, err := h.forecastingSvc.ListPatternExclusions(r.Context(), userID)
	if err != nil {
		h.Logger.Error("Failed to list pattern exclusions", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to list exclusions")
		return
	}

	writeJSON(w, http.StatusOK, exclusions)
}
