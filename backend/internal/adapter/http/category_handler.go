package http

import (
	"encoding/json"
	"net/http"

	"cogni-cash/internal/domain/entity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type categoryRequest struct {
	Name               string `json:"name"`
	Color              string `json:"color"`
	IsVariableSpending bool   `json:"is_variable_spending"`
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	if h.categoryRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "category repository not available")
		return
	}
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cats, err := h.categoryRepo.FindAll(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cats == nil {
		cats = []entity.Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	if h.categoryRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "category repository not available")
		return
	}
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid request body: 'name' is required")
		return
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	cat, err := h.categoryRepo.Save(r.Context(), entity.Category{
		UserID:             userID,
		Name:               req.Name,
		Color:              req.Color,
		IsVariableSpending: req.IsVariableSpending,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	if h.categoryRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "category repository not available")
		return
	}
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid request body: 'name' is required")
		return
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	cat, err := h.categoryRepo.Update(r.Context(), entity.Category{
		ID:                 id,
		UserID:             userID,
		Name:               req.Name,
		Color:              req.Color,
		IsVariableSpending: req.IsVariableSpending,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	if h.categoryRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "category repository not available")
		return
	}
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}
	if err := h.categoryRepo.Delete(r.Context(), id, userID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) restoreCategory(w http.ResponseWriter, r *http.Request) {
	if h.categoryRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "category repository not available")
		return
	}
	userID := h.getUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	// Fetch the category to preserve existing name/color
	cat, err := h.categoryRepo.FindByID(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	cat.DeletedAt = nil // Clear deleted_at
	restored, err := h.categoryRepo.Update(r.Context(), cat)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, restored)
}
