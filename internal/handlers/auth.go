package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
)

type AuthHandler struct {
	service services.AuthServiceI
	log     *zap.SugaredLogger
}

func NewAuthHandler(service services.AuthServiceI, log *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{service: service, log: log}
}

// Register
// @Summary Register
// @Tags System — Auth
// @Accept json
// @Produce json
// @Param data body models.RegisterRequest true "register payload"
// @Success 200 {object} models.User
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 409 {object} helpers.ErrorData "Email уже зарегистрирован"
// @Failure 500 {object} helpers.ErrorData "Ошибка сервера"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Ошибка декодирования JSON при регистрации", "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch {
		case helpers.IsInvalidInput(err):
			helpers.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, services.ErrEmailTaken):
			helpers.Error(w, http.StatusConflict, "Пользователь с таким email уже существует")
		default:
			h.log.Errorw("Ошибка регистрации", "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка регистрации")
		}
		return
	}

	h.log.Infow("Пользователь успешно зарегистрирован", "email", req.Email, "id", user.ID)
	helpers.JSON(w, http.StatusOK, user)
}

// Login
// @Summary Login
// @Tags System — Auth
// @Accept json
// @Produce json
// @Param data body models.LoginRequest true "login payload"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} helpers.ErrorData "Некорректный запрос"
// @Failure 401 {object} helpers.ErrorData "Неверный email или пароль"
// @Failure 500 {object} helpers.ErrorData "Ошибка сервера"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Ошибка декодирования JSON при логине", "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный запрос")
		return
	}

	token, err := h.service.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			h.log.Warnw("Неудачная попытка входа", "email", req.Email)
			helpers.Error(w, http.StatusUnauthorized, "Неверный email или пароль")
		default:
			h.log.Errorw("Ошибка логина", "email", req.Email, "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка входа")
		}
		return
	}

	h.log.Infow("Пользователь вошёл в систему", "email", req.Email)
	helpers.JSON(w, http.StatusOK, models.AuthResponse{Token: token})
}
