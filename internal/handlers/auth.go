package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
)

type AuthHandler struct {
	service *services.AuthService
	log     *zap.SugaredLogger
}

func NewAuthHandler(service *services.AuthService, log *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{service: service, log: log}
}

// Register
// @Summary Register
// @Tags auth
// @Accept json
// @Produce json
// @Param data body models.RegisterRequest true "register payload"
// @Success 200 {object} models.User
// @Failure 400 {object} helpers.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}
	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	helpers.JSON(w, http.StatusOK, user)
}

// Login
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param data body models.LoginRequest true "login payload"
// @Success 200 {object} models.AuthResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}
	token, err := h.service.Login(r.Context(), req)
	if err != nil {
		helpers.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	helpers.JSON(w, http.StatusOK, models.AuthResponse{Token: token})
}

// Me
// @Summary Me
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 200 {string} string "You are authorized"
// @Failure 401 {object} helpers.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("You are authorized"))
}
