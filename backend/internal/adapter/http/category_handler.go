package http

import (
	"encoding/json"
	"net/http"

	"cogni-cash/internal/domain/entity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"cogni-cash/internal/domain/port"
)

type CategoryHandler struct {
	categorySvc port.CategoryUseCase
	forecastingSvc port.ForecastingUseCase
}

func NewCategoryHandler(categorySvc port.CategoryUseCase, forecastingSvc port.ForecastingUseCase) *CategoryHandler {
	return &CategoryHandler{
		categorySvc: categorySvc,
		forecastingSvc: forecastingSvc,
	}
}

type categoryRequest struct {
	Name               string `json:"name"`
	Color              string `json:"color"`
	IsVariableSpending bool   `json:"is_variable_spending"`
	ForecastStrategy   string `json:"forecast_strategy"`
}

type shareCategoryRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	Permission string    `json:"permission"` // 'view' or 'edit'
}

func (h *CategoryHandler) listCategories(w http.ResponseWriter, r *http.Request) {
	if h.categorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "category service not available")
		return
	}
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cats, err := h.categorySvc.GetAll(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cats == nil {
		cats = []entity.Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *CategoryHandler) createCategory(w http.ResponseWriter, r *http.Request) {
	if h.categorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "category service not available")
		return
	}
	userID := GetUserID(r.Context())
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
	cat, err := h.categorySvc.Create(r.Context(), entity.Category{
		UserID:             userID,
		Name:               req.Name,
		Color:              req.Color,
		IsVariableSpending: req.IsVariableSpending,
		ForecastStrategy:   req.ForecastStrategy,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

func (h *CategoryHandler) updateCategory(w http.ResponseWriter, r *http.Request) {
	if h.categorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "category service not available")
		return
	}
	userID := GetUserID(r.Context())
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
	cat, err := h.categorySvc.Update(r.Context(), entity.Category{
		ID:                 id,
		UserID:             userID,
		Name:               req.Name,
		Color:              req.Color,
		IsVariableSpending: req.IsVariableSpending,
		ForecastStrategy:   req.ForecastStrategy,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *CategoryHandler) getCategoryAverage(w http.ResponseWriter, r *http.Request) {
	if h.forecastingSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "forecasting service not available")
		return
	}
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	strategy := r.URL.Query().Get("strategy")
	if strategy == "" {
		strategy = "3y"
	}

	avg, err := h.forecastingSvc.CalculateCategoryAverage(r.Context(), userID, id, strategy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]float64{"average": avg})
}

func (h *CategoryHandler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	if h.categorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "category service not available")
		return
	}
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}
	if err := h.categorySvc.Delete(r.Context(), id, userID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) restoreCategory(w http.ResponseWriter, r *http.Request) {
	if h.categorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "category service not available")
		return
	}
	userID := GetUserID(r.Context())
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
	cat, err := h.categorySvc.GetByID(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	cat.DeletedAt = nil // Clear deleted_at
	restored, err := h.categorySvc.Update(r.Context(), cat)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, restored)
}

func (h *CategoryHandler) shareCategory(w http.ResponseWriter, r *http.Request) {
	if h.categorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "category service not available")
		return
	}
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	var req shareCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Permission == "" {
		req.Permission = "view"
	}

	err = h.categorySvc.ShareCategory(r.Context(), id, userID, req.UserID, req.Permission)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *CategoryHandler) revokeCategoryShare(w http.ResponseWriter, r *http.Request) {
	if h.categorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "category service not available")
		return
	}
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	categoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	sharedWithUserID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	err = h.categorySvc.RevokeShare(r.Context(), categoryID, userID, sharedWithUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) listCategoryShares(w http.ResponseWriter, r *http.Request) {
	if h.categorySvc == nil {
		writeError(w, http.StatusServiceUnavailable, "category service not available")
		return
	}
	userID := GetUserID(r.Context())
	if userID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	categoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid category id")
		return
	}

	// Verify the user has access to this category before listing its shares
	_, err = h.categorySvc.GetByID(r.Context(), categoryID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	shares, err := h.categorySvc.ListShares(r.Context(), categoryID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if shares == nil {
		shares = []uuid.UUID{}
	}
	writeJSON(w, http.StatusOK, shares)
}
