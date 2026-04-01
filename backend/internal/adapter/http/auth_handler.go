package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// maxFieldLen is the maximum byte length accepted for username / password
// fields. bcrypt silently truncates at 72 bytes, so we reject anything
// obviously oversized before it reaches the crypto layer.
const maxFieldLen = 256

const authTokenCookieName = "auth_token"

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordConfirmRequest struct {
	Token       string `json:"token"`
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

	// Set HttpOnly cookie for web clients
	http.SetCookie(w, &http.Cookie{
		Name:     authTokenCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set Secure only if request is HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// Set non-HttpOnly cookie for UI session tracking
	http.SetCookie(w, &http.Cookie{
		Name:     "cogni_cash_logged_in",
		Value:    "true",
		Path:     "/",
		HttpOnly: false,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	h.Logger.Info("Successful login", "username", req.Username)
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     authTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "cogni_cash_logged_in",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
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

func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Email) == 0 {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	if err := h.authSvc.RequestPasswordReset(r.Context(), req.Email); err != nil {
		h.Logger.Error("Forgot password request failed", "email", req.Email, "error", err)
		// We still return 200 to prevent enumeration
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "If an account exists with this email, a reset link has been sent."})
}

func (h *Handler) validateResetToken(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	isValid, err := h.authSvc.ValidateResetToken(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to validate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"valid": isValid})
}

func (h *Handler) confirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Token) == 0 || len(req.NewPassword) == 0 {
		writeError(w, http.StatusBadRequest, "token and new_password are required")
		return
	}

	if err := h.authSvc.ConfirmPasswordReset(r.Context(), req.Token, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string
		authHeader := r.Header.Get("Authorization")

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Fallback to cookie if Authorization header is missing
			cookie, err := r.Cookie(authTokenCookieName)
			if err == nil {
				tokenStr = cookie.Value
			}
		}

		if tokenStr == "" {
			writeError(w, http.StatusUnauthorized, "missing or invalid authorization")
			return
		}

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
