package http

import (
	"net/http"

	"github.com/google/uuid"
)

// GET /api/v1/sharing/dashboard/
func (h *Handler) getSharingDashboard(w http.ResponseWriter, r *http.Request) {
	if h.sharingSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "sharing service not available")
		return
	}

	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dashboard, err := h.sharingSvc.GetDashboard(r.Context(), userID)
	if err != nil {
		h.Logger.Error("Failed to fetch sharing dashboard", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "failed to fetch sharing dashboard")
		return
	}

	writeJSON(w, http.StatusOK, dashboard)
}
