package http

import (
	"encoding/json"
	"net/http"

	"cogni-cash/internal/domain/entity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createUserRequest struct {
	entity.User
	Password string `json:"password"`
}

// getMe returns the currently authenticated user's profile
func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		h.Logger.Warn("getMe failed: missing or invalid userID in context")
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userSvc.GetUser(r.Context(), userID.String())
	if err != nil {
		h.Logger.Error("getMe failed: could not fetch profile", "user_id", userID, "error", err)
		writeError(w, http.StatusNotFound, "user profile not found")
		return
	}

	h.Logger.Info("getMe successful", "user_id", userID, "username", user.Username)
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	h.Logger.Info("Fetching user list", "search_query", search)

	users, err := h.userSvc.ListUsers(r.Context(), search)
	if err != nil {
		h.Logger.Error("listUsers failed", "search_query", search, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	if users == nil {
		users = []entity.User{}
	}

	h.Logger.Info("listUsers successful", "count", len(users))
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.Logger.Info("Fetching specific user", "target_id", id)

	user, err := h.userSvc.GetUser(r.Context(), id)
	if err != nil {
		h.Logger.Warn("getUser failed: user not found", "target_id", id, "error", err)
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Warn("createUser failed: invalid JSON body", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.Logger.Info("Attempting to create new user", "username", req.Username, "email", req.Email, "role", req.Role)

	user, err := h.userSvc.CreateUser(r.Context(), req.User, req.Password)
	if err != nil {
		h.Logger.Error("createUser failed: domain service error", "username", req.Username, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create user - check if username or email already exists")
		return
	}

	h.Logger.Info("createUser successful", "new_user_id", user.ID, "username", user.Username)

	// Send welcome email asynchronously
	if h.notificationSvc != nil && h.WaitGroup != nil {
		h.WaitGroup.Add(1)
		go func() {
			defer h.WaitGroup.Done()
			// Use AppCtx so the task is cancelled on shutdown, or finishes before exit
			if err := h.notificationSvc.SendWelcomeEmail(h.AppCtx, user); err != nil {
				h.Logger.Error("Failed to send welcome email in background", "user_id", user.ID, "error", err)
			}
		}()
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.Logger.Info("Attempting to update user", "target_id", id)

	var req entity.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Warn("updateUser failed: invalid JSON body", "target_id", id, "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userSvc.UpdateUser(r.Context(), id, req)
	if err != nil {
		h.Logger.Error("updateUser failed: domain service error", "target_id", id, "error", err)
		writeError(w, http.StatusBadRequest, "failed to update user")
		return
	}

	h.Logger.Info("updateUser successful", "target_id", id, "updated_username", user.Username)
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	currentUserID := userID.String()

	h.Logger.Info("Attempting to delete user", "target_id", id, "requested_by", currentUserID)

	if id == currentUserID {
		h.Logger.Warn("deleteUser failed: user attempted to delete themselves", "user_id", currentUserID)
		writeError(w, http.StatusBadRequest, "you cannot delete your own account")
		return
	}

	if err := h.userSvc.DeleteUser(r.Context(), id); err != nil {
		h.Logger.Error("deleteUser failed: domain service error", "target_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	h.Logger.Info("deleteUser successful", "deleted_user_id", id)
	w.WriteHeader(http.StatusNoContent)
}
