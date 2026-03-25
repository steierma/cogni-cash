package http

import (
	"encoding/json"
	"net/http"

	"cogni-cash/internal/domain/entity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type categoryRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	if h.categoryRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "category repository not available")
		return
	}
	cats, err := h.categoryRepo.FindAll(r.Context())
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
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid request body: 'name' is required")
		return
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	cat, err := h.categoryRepo.Save(r.Context(), entity.Category{Name: req.Name, Color: req.Color})
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
	cat, err := h.categoryRepo.Update(r.Context(), entity.Category{ID: id, Name: req.Name, Color: req.Color})
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
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}
	if err := h.categoryRepo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
