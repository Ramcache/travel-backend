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

type ProfileHandler struct {
	service services.AuthServiceI
	log     *zap.SugaredLogger
}

func NewProfileHandler(s services.AuthServiceI, log *zap.SugaredLogger) *ProfileHandler {
	return &ProfileHandler{service: s, log: log}
}

// Get profile
// @Summary Получить профиль текущего пользователя
// @Security Bearer
// @Tags profile
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} helpers.ErrorData "Неавторизованный доступ"
// @Failure 404 {object} helpers.ErrorData "Пользователь не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при загрузке профиля"
// @Router /profile [get]
func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := helpers.GetUserID(r.Context())

	u, err := h.service.GetByID(r.Context(), uid)
	switch {
	case errors.Is(err, services.ErrNotFound):
		h.log.Warnw("Профиль пользователя не найден", "uid", uid)
		helpers.Error(w, http.StatusNotFound, "Пользователь не найден")
		return
	case err != nil:
		h.log.Errorw("Ошибка загрузки профиля", "uid", uid, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось загрузить профиль")
		return
	}

	h.log.Infow("Профиль успешно загружен", "uid", uid)
	helpers.JSON(w, http.StatusOK, u)
}

// Update profile
// @Summary Обновить профиль текущего пользователя
// @Security Bearer
// @Tags profile
// @Accept json
// @Produce json
// @Param body body models.UpdateProfileRequest true "Данные профиля"
// @Success 200 {object} models.User
// @Failure 400 {object} helpers.ErrorData "Некорректное тело запроса"
// @Failure 401 {object} helpers.ErrorData "Неавторизованный доступ"
// @Failure 404 {object} helpers.ErrorData "Пользователь не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при обновлении профиля"
// @Router /profile [put]
func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := helpers.GetUserID(r.Context())
	if uid == 0 {
		h.log.Warn("Попытка обновления профиля без авторизации")
		helpers.Error(w, http.StatusUnauthorized, "Неавторизованный доступ")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Некорректный JSON при обновлении профиля", "uid", uid, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	u, err := h.service.UpdateProfile(r.Context(), uid, req)
	switch {
	case errors.Is(err, services.ErrNotFound):
		h.log.Warnw("Пользователь не найден для обновления профиля", "uid", uid)
		helpers.Error(w, http.StatusNotFound, "Пользователь не найден")
		return
	case helpers.IsInvalidInput(err):
		h.log.Warnw("Ошибка валидации при обновлении профиля", "uid", uid, "err", err)
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка обновления профиля", "uid", uid, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить профиль")
		return
	}

	h.log.Infow("Профиль успешно обновлён", "uid", uid)
	helpers.JSON(w, http.StatusOK, u)
}
