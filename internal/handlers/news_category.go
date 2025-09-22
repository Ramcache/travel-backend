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
	"github.com/Ramcache/travel-backend/internal/services"
)

type NewsCategoryHandler struct {
	service *services.NewsCategoryService
	log     *zap.SugaredLogger
}

func NewNewsCategoryHandler(s *services.NewsCategoryService, log *zap.SugaredLogger) *NewsCategoryHandler {
	return &NewsCategoryHandler{service: s, log: log}
}

// List
// @Summary Получить список категорий новостей
// @Tags news-categories
// @Produce json
// @Success 200 {array} models.NewsCategory
// @Failure 500 {object} helpers.ErrorData "Не удалось получить категории"
// @Router /admin/news/categories [get]
func (h *NewsCategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.List(r.Context())
	if err != nil {
		h.log.Errorw("Ошибка получения категорий", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить категории")
		return
	}
	helpers.JSON(w, http.StatusOK, list)
}

// Get
// @Summary Получить категорию по ID
// @Tags news-categories
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} models.NewsCategory
// @Failure 404 {object} helpers.ErrorData "Категория не найдена"
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении категории"
// @Router /admin/news/categories/{id} [get]
func (h *NewsCategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	c, err := h.service.GetByID(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrCategoryNotFound):
		helpers.Error(w, http.StatusNotFound, "Категория не найдена")
		return
	case err != nil:
		h.log.Errorw("Ошибка получения категории", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Ошибка при получении категории")
		return
	}
	helpers.JSON(w, http.StatusOK, c)
}

// Create
// @Summary Создать категорию новостей
// @Tags news-categories
// @Accept json
// @Produce json
// @Param data body models.CreateNewsCategoryRequest true "Данные категории"
// @Success 201 {object} models.NewsCategory
// @Failure 400 {object} helpers.ErrorData "Некорректное тело запроса или ошибка валидации"
// @Failure 500 {object} helpers.ErrorData "Не удалось создать категорию"
// @Router /admin/news/categories [post]
func (h *NewsCategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNewsCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}
	c, err := h.service.Create(r.Context(), req)
	switch {
	case helpers.IsInvalidInput(err):
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка создания категории", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось создать категорию")
		return
	}
	helpers.JSON(w, http.StatusCreated, c)
}

// Update
// @Summary Обновить категорию новостей
// @Tags news-categories
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Param data body models.UpdateNewsCategoryRequest true "Данные для обновления"
// @Success 200 {object} models.NewsCategory
// @Failure 400 {object} helpers.ErrorData "Некорректное тело запроса или ошибка валидации"
// @Failure 404 {object} helpers.ErrorData "Категория не найдена"
// @Failure 500 {object} helpers.ErrorData "Не удалось обновить категорию"
// @Router /admin/news/categories/{id} [put]
func (h *NewsCategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.UpdateNewsCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	c, err := h.service.Update(r.Context(), id, req)
	switch {
	case errors.Is(err, services.ErrCategoryNotFound):
		helpers.Error(w, http.StatusNotFound, "Категория не найдена")
		return
	case helpers.IsInvalidInput(err):
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка обновления категории", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить категорию")
		return
	}
	helpers.JSON(w, http.StatusOK, c)
}

// Delete
// @Summary Удалить категорию новостей
// @Tags news-categories
// @Param id path int true "ID категории"
// @Success 204 "Категория успешно удалена"
// @Failure 404 {object} helpers.ErrorData "Категория не найдена"
// @Failure 500 {object} helpers.ErrorData "Не удалось удалить категорию"
// @Router /admin/news/categories/{id} [delete]
func (h *NewsCategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	err := h.service.Delete(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrCategoryNotFound):
		helpers.Error(w, http.StatusNotFound, "Категория не найдена")
		return
	case err != nil:
		h.log.Errorw("Ошибка удаления категории", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить категорию")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
