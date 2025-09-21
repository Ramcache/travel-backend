package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/middleware"
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
// @Tags news
// @Produce json
// @Param category query string false "Фильтр: hadj|company"
// @Param media_type query string false "Тип медиа: photo|video"
// @Param search query string false "Поиск по заголовку или excerpt"
// @Param page query int false "Номер страницы (1)"
// @Param limit query int false "Размер страницы (12)"
// @Success 200 {object} map[string]interface{}
// @Router /news [get]
func (h *NewsHandler) PublicList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	items, total, err := h.service.ListPublic(r.Context(), models.ListNewsParams{
		Category:  q.Get("category"),
		MediaType: q.Get("media_type"),
		Search:    q.Get("search"),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		h.log.Errorw("news_list_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "failed to list news")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"meta":  map[string]int{"total": total, "page": ifZero(page, 1), "limit": ifZero(limit, 12)},
	})
}

// PublicGet
// @Summary Get news by slug or id (public)
// @Tags news
// @Produce json
// @Param slug_or_id path string true "Slug или ID новости"
// @Param id path int true "ID новости"
// @Success 200 {object} models.News
// @Failure 404 {object} helpers.ErrorResponse
// @Router /news/{slug_or_id} [get]
func (h *NewsHandler) PublicGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slug_or_id")
	n, err := h.service.GetPublic(r.Context(), id)
	if err != nil {
		h.log.Errorw("news_get_failed", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "failed to get news")
		return
	}
	if n == nil {
		helpers.Error(w, http.StatusNotFound, "news not found")
		return
	}
	helpers.JSON(w, http.StatusOK, n)
}

// AdminList
// @Summary List news (admin)
// @Security Bearer
// @Tags admin-news
// @Produce json
// @Param status query string false "Статус: draft|published|archived"
// @Param category query string false "Фильтр по категории"
// @Param media_type query string false "Фильтр по типу медиа"
// @Param search query string false "Поиск по заголовку/анонсу"
// @Param page query int false "Номер страницы"
// @Param limit query int false "Размер страницы (по умолчанию 20)"
// @Success 200 {object} map[string]interface{}
// @Router /admin/news [get]
func (h *NewsHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	items, total, err := h.service.AdminList(r.Context(), models.ListNewsParams{
		Category: q.Get("category"), MediaType: q.Get("media_type"), Search: q.Get("search"), Status: q.Get("status"), Page: page, Limit: limit,
	})
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to list news")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]interface{}{"items": items, "meta": map[string]int{"total": total, "page": ifZero(page, 1), "limit": ifZero(limit, 20)}})
}

// Create
// @Summary Create news (admin)
// @Security Bearer
// @Tags admin-news
// @Accept json
// @Produce json
// @Param body body models.CreateNewsRequest true "payload"
// @Success 201 {object} models.News
// @Router /admin/news [post]
func (h *NewsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	var authorID *int
	if v := r.Context().Value(middleware.UserIDKey); v != nil {
		if id, ok := v.(int); ok {
			authorID = &id
		}
	}
	n, err := h.service.Create(r.Context(), authorID, req)
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	helpers.JSON(w, http.StatusCreated, n)
}

// Update
// @Summary Update news (admin)
// @Security Bearer
// @Tags admin-news
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param body body models.UpdateNewsRequest true "payload"
// @Success 200 {object} models.News
// @Router /admin/news/{id} [put]
func (h *NewsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.UpdateNewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	n, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if n == nil {
		helpers.Error(w, http.StatusNotFound, "news not found")
		return
	}
	helpers.JSON(w, http.StatusOK, n)
}

// Delete
// @Summary Delete news (admin)
// @Security Bearer
// @Tags admin-news
// @Param id path int true "id"
// @Success 204 {string} string ""
// @Router /admin/news/{id} [delete]
func (h *NewsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.service.Delete(r.Context(), id); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ifZero(v, d int) int {
	if v == 0 {
		return d
	}
	return v
}
