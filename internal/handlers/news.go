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

type NewsHandler struct {
	service *services.NewsService
	log     *zap.SugaredLogger
}

func NewNewsHandler(s *services.NewsService, log *zap.SugaredLogger) *NewsHandler {
	return &NewsHandler{service: s, log: log}
}

// PublicList
// @Summary List news (public)
// @Description Публичный список новостей с фильтрами и пагинацией
// @Tags Public — News
// @Produce json
// @Param category_id query string false "Фильтр по категории"
// @Param media_type query string false "Тип медиа: photo|video"
// @Param search query string false "Поиск по заголовку или анонсу"
// @Param page query int false "Номер страницы (1)"
// @Param limit query int false "Размер страницы (12)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} helpers.ErrorData "Не удалось получить список новостей"
// @Router /news [get]
func (h *NewsHandler) PublicList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	catID, _ := strconv.Atoi(q.Get("category_id"))

	items, total, err := h.service.ListPublic(r.Context(), models.ListNewsParams{
		CategoryID: catID,
		MediaType:  q.Get("media_type"),
		Search:     q.Get("search"),
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		h.log.Errorw("Ошибка получения списка новостей", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить список новостей")
		return
	}

	h.log.Infow("Список новостей успешно получен", "total", total)
	helpers.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"meta":  map[string]int{"total": total, "page": ifZero(page, 1), "limit": ifZero(limit, 12)},
	})
}

// PublicGet
// @Summary Get news by slug or id (public)
// @Tags Public — News
// @Produce json
// @Param slug_or_id path string true "Slug или ID новости"
// @Success 200 {object} models.News
// @Failure 404 {object} helpers.ErrorData "Новость не найдена"
// @Failure 500 {object} helpers.ErrorData "Не удалось получить новость"
// @Router /news/{slug_or_id} [get]
func (h *NewsHandler) PublicGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slug_or_id")

	n, err := h.service.GetPublic(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrNotFound):
		h.log.Warnw("Новость не найдена", "id", id)
		helpers.Error(w, http.StatusNotFound, "Новость не найдена")
		return
	case err != nil:
		h.log.Errorw("Ошибка получения новости", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить новость")
		return
	}

	h.log.Infow("Новость успешно получена", "id", id)
	helpers.JSON(w, http.StatusOK, n)
}

// AdminList
// @Summary List news (admin)
// @Security Bearer
// @Tags Admin — News
// @Produce json
// @Param status query string false "Статус: draft|published|archived"
// @Param category_id query string false "Фильтр по категории"
// @Param media_type query string false "Фильтр по типу медиа"
// @Param search query string false "Поиск по заголовку/анонсу"
// @Param page query int false "Номер страницы"
// @Param limit query int false "Размер страницы (по умолчанию 20)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} helpers.ErrorData "Не удалось получить список новостей"
// @Router /admin/news [get]
func (h *NewsHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	catID, _ := strconv.Atoi(q.Get("category_id"))

	items, total, err := h.service.AdminList(r.Context(), models.ListNewsParams{
		CategoryID: catID,
		MediaType:  q.Get("media_type"),
		Search:     q.Get("search"),
		Status:     q.Get("status"),
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		h.log.Errorw("Ошибка получения списка новостей (admin)", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить список новостей")
		return
	}

	h.log.Infow("Список новостей (admin) успешно получен", "total", total)
	helpers.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"meta":  map[string]int{"total": total, "page": ifZero(page, 1), "limit": ifZero(limit, 20)},
	})
}

// Create
// @Summary Create news (admin)
// @Security Bearer
// @Tags Admin — News
// @Accept json
// @Produce json
// @Param body body models.CreateNewsRequest true "payload"
// @Success 201 {object} models.News
// @Failure 400 {object} helpers.ErrorData "Некорректный JSON или ошибка валидации"
// @Failure 500 {object} helpers.ErrorData "Ошибка сервера"
// @Router /admin/news [post]
func (h *NewsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Ошибка парсинга JSON при создании новости", "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	var authorID *int
	if v := r.Context().Value(helpers.UserIDKey); v != nil {
		if id, ok := v.(int); ok {
			authorID = &id
		}
	}

	n, err := h.service.Create(r.Context(), authorID, req)
	switch {
	case helpers.IsInvalidInput(err):
		h.log.Warnw("Ошибка валидации при создании новости", "err", err)
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка создания новости", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Ошибка создания новости")
		return
	}

	h.log.Infow("Новость успешно создана", "id", n.ID)
	helpers.JSON(w, http.StatusCreated, n)
}

// Update
// @Summary Update news (admin)
// @Security Bearer
// @Tags Admin — News
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param body body models.UpdateNewsRequest true "payload"
// @Success 200 {object} models.News
// @Failure 400 {object} helpers.ErrorData "Некорректный JSON или ошибка валидации"
// @Failure 404 {object} helpers.ErrorData "Новость не найдена"
// @Failure 500 {object} helpers.ErrorData "Ошибка сервера"
// @Router /admin/news/{id} [put]
func (h *NewsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.UpdateNewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Ошибка парсинга JSON при обновлении новости", "id", id, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	n, err := h.service.Update(r.Context(), id, req)
	switch {
	case errors.Is(err, services.ErrNotFound):
		h.log.Warnw("Новость не найдена для обновления", "id", id)
		helpers.Error(w, http.StatusNotFound, "Новость не найдена")
		return
	case helpers.IsInvalidInput(err):
		h.log.Warnw("Ошибка валидации при обновлении новости", "id", id, "err", err)
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка обновления новости", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Ошибка обновления новости")
		return
	}

	h.log.Infow("Новость успешно обновлена", "id", id)
	helpers.JSON(w, http.StatusOK, n)
}

// Delete
// @Summary Delete news (admin)
// @Security Bearer
// @Tags Admin — News
// @Param id path int true "id"
// @Success 204 {string} string ""
// @Failure 404 {object} helpers.ErrorData "Новость не найдена"
// @Failure 500 {object} helpers.ErrorData "Ошибка сервера"
// @Router /admin/news/{id} [delete]
func (h *NewsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	err := h.service.Delete(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrNotFound):
		h.log.Warnw("Новость не найдена для удаления", "id", id)
		helpers.Error(w, http.StatusNotFound, "Новость не найдена")
		return
	case err != nil:
		h.log.Errorw("Ошибка удаления новости", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить новость")
		return
	}

	h.log.Infow("Новость успешно удалена", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

// Recent
// @Summary Get recent news
// @Tags Public — News
// @Produce json
// @Param limit query int false "Количество новостей (по умолчанию 3)"
// @Success 200 {array} models.News
// @Failure 500 {object} helpers.ErrorData "Не удалось получить последние новости"
// @Router /news/recent [get]
func (h *NewsHandler) Recent(w http.ResponseWriter, r *http.Request) {
	limit := 3
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}

	news, err := h.service.GetRecent(r.Context(), limit)
	if err != nil {
		h.log.Errorw("Ошибка получения последних новостей", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить последние новости")
		return
	}

	h.log.Infow("Последние новости успешно получены", "count", len(news))
	helpers.JSON(w, http.StatusOK, news)
}

// Popular
// @Summary Get popular news
// @Tags Public — News
// @Produce json
// @Param limit query int false "Количество новостей (по умолчанию 5)"
// @Success 200 {array} models.News
// @Failure 500 {object} helpers.ErrorData "Не удалось получить популярные новости"
// @Router /news/popular [get]
func (h *NewsHandler) Popular(w http.ResponseWriter, r *http.Request) {
	limit := 5
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}

	news, err := h.service.GetPopular(r.Context(), limit)
	if err != nil {
		h.log.Errorw("Ошибка получения популярных новостей", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить популярные новости")
		return
	}

	h.log.Infow("Популярные новости успешно получены", "count", len(news))
	helpers.JSON(w, http.StatusOK, news)
}

func ifZero(v, d int) int {
	if v == 0 {
		return d
	}
	return v
}
