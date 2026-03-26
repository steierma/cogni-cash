package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// maxFieldLen is the maximum byte length accepted for username / password
// fields. bcrypt silently truncates at 72 bytes, so we reject anything
// obviously oversized before it reaches the crypto layer.
const maxFieldLen = 256

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Warn("Invalid login request body", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Username) == 0 || len(req.Password) == 0 {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}
	if len(req.Username) > maxFieldLen || len(req.Password) > maxFieldLen {
		writeError(w, http.StatusBadRequest, "username or password exceeds maximum length")
		return
	}

	token, err := h.authSvc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	h.Logger.Info("Successful login", "username", req.Username)
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.OldPassword) == 0 || len(req.NewPassword) == 0 {
		writeError(w, http.StatusBadRequest, "old_password and new_password are required")
		return
	}
	if len(req.OldPassword) > maxFieldLen || len(req.NewPassword) > maxFieldLen {
		writeError(w, http.StatusBadRequest, "password exceeds maximum length")
		return
	}

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.authSvc.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := h.authSvc.ValidateToken(tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr, ok := r.Context().Value(userIDKey).(string)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := h.userSvc.GetUser(r.Context(), userIDStr)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to verify user permissions")
			return
		}

		if user.Role != "admin" {
			h.Logger.Warn("Forbidden action attempted by non-admin", "user_id", userIDStr, "path", r.URL.Path)
			writeError(w, http.StatusForbidden, "forbidden: administrator access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}
