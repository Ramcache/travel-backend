package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type UserHandler struct {
	repo repository.UserRepoI
	log  *zap.SugaredLogger
}

func NewUserHandler(repo repository.UserRepoI, log *zap.SugaredLogger) *UserHandler {
	return &UserHandler{repo: repo, log: log}
}

// List
// @Summary Получить всех пользователей
// @Tags Admin — Users
// @Security Bearer
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} helpers.ErrorData "Не удалось получить список пользователей"
// @Router /admin/users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.GetAll(r.Context())
	if err != nil {
		h.log.Errorw("Ошибка получения списка пользователей", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить список пользователей")
		return
	}

	h.log.Infow("Список пользователей успешно получен", "count", len(users))
	helpers.JSON(w, http.StatusOK, users)
}

// Get
// @Summary Получить пользователя по ID
// @Tags Admin — Users
// @Security Bearer
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 404 {object} helpers.ErrorData "Пользователь не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении пользователя"
// @Router /admin/users/{id} [get]
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	user, err := h.repo.GetByID(r.Context(), id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		h.log.Warnw("Пользователь не найден", "id", id)
		helpers.Error(w, http.StatusNotFound, "Пользователь не найден")
		return
	case err != nil:
		h.log.Errorw("Ошибка получения пользователя", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить пользователя")
		return
	}

	h.log.Infow("Пользователь успешно получен", "id", id)
	helpers.JSON(w, http.StatusOK, user)
}

// Create
// @Summary Создать пользователя
// @Tags Admin — Users
// @Security Bearer
// @Accept json
// @Produce json
// @Param data body models.CreateUserRequest true "User data"
// @Success 200 {object} models.User
// @Failure 400 {object} helpers.ErrorData "Некорректное тело запроса"
// @Failure 500 {object} helpers.ErrorData "Ошибка при создании пользователя"
// @Router /admin/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Некорректный JSON при создании пользователя", "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	hash, err := helpers.HashPassword(req.Password)
	if err != nil {
		h.log.Errorw("Ошибка хэширования пароля", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обработать пароль")
		return
	}

	user := &models.User{
		Email:    req.Email,
		Password: hash,
		FullName: req.FullName,
		RoleID:   req.RoleID,
	}

	if err := h.repo.Create(r.Context(), user); err != nil {
		h.log.Errorw("Ошибка создания пользователя", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось создать пользователя")
		return
	}

	h.log.Infow("Пользователь успешно создан", "id", user.ID, "email", user.Email)
	helpers.JSON(w, http.StatusOK, user)
}

// Update
// @Summary Обновить данные пользователя
// @Tags Admin — Users
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param data body models.UpdateUserRequest true "User update"
// @Success 200 {object} models.User
// @Failure 400 {object} helpers.ErrorData "Некорректное тело запроса"
// @Failure 404 {object} helpers.ErrorData "Пользователь не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при обновлении пользователя"
// @Router /admin/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Некорректный JSON при обновлении пользователя", "id", id, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	user, err := h.repo.GetByID(r.Context(), id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		h.log.Warnw("Пользователь не найден для обновления", "id", id)
		helpers.Error(w, http.StatusNotFound, "Пользователь не найден")
		return
	case err != nil:
		h.log.Errorw("Ошибка поиска пользователя для обновления", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить пользователя")
		return
	}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.RoleID != nil {
		user.RoleID = *req.RoleID
	}

	if err := h.repo.Update(r.Context(), user); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.log.Warnw("Пользователь не найден при обновлении", "id", id)
			helpers.Error(w, http.StatusNotFound, "Пользователь не найден")
		default:
			h.log.Errorw("Ошибка обновления пользователя", "id", id, "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить пользователя")
		}
		return
	}

	h.log.Infow("Пользователь успешно обновлён", "id", id)
	helpers.JSON(w, http.StatusOK, user)
}

// Delete
// @Summary Удалить пользователя
// @Tags Admin — Users
// @Security Bearer
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 404 {object} helpers.ErrorData "Пользователь не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при удалении пользователя"
// @Router /admin/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	err := h.repo.Delete(r.Context(), id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		h.log.Warnw("Пользователь не найден для удаления", "id", id)
		helpers.Error(w, http.StatusNotFound, "Пользователь не найден")
		return
	case err != nil:
		h.log.Errorw("Ошибка удаления пользователя", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить пользователя")
		return
	}

	h.log.Infow("Пользователь успешно удалён", "id", id)
	w.WriteHeader(http.StatusNoContent)
}
