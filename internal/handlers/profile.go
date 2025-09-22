package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
)

type ProfileHandler struct {
	service *services.AuthService
}

func NewProfileHandler(s *services.AuthService) *ProfileHandler {
	return &ProfileHandler{service: s}
}

// Get profile
// @Summary Get current user profile
// @Security Bearer
// @Tags profile
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} helpers.ErrorResponse
// @Router /profile [get]
func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := helpers.GetUserID(r.Context())
	if uid == 0 {
		helpers.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	u, err := h.service.GetByID(r.Context(), uid)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to load profile")
		return
	}
	if u == nil {
		helpers.Error(w, http.StatusNotFound, "user not found")
		return
	}
	helpers.JSON(w, http.StatusOK, u)
}

// Update profile
// @Summary Update current user profile
// @Security Bearer
// @Tags profile
// @Accept json
// @Produce json
// @Param body body models.UpdateProfileRequest true "profile"
// @Success 200 {object} models.User
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Router /profile [put]
func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := helpers.GetUserID(r.Context())
	if uid == 0 {
		helpers.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.service.UpdateProfile(r.Context(), uid, req)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	helpers.JSON(w, http.StatusOK, u)
}
