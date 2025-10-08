package handlers

import (
	"errors"
	"github.com/Ramcache/travel-backend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/services"
)

type TripPageHandler struct {
	svc *services.TripPageService
	log *zap.SugaredLogger
}

func NewTripPageHandler(svc *services.TripPageService, log *zap.SugaredLogger) *TripPageHandler {
	return &TripPageHandler{svc: svc, log: log}
}

// Get
// @Summary Trip page data
// @Description Полный набор данных для страницы тура (тур, отели, отзывы, популярные туры, новости, курсы, countdown)
// @Tags Public — Trips
// @Produce json
// @Param id path int true "Trip ID"
// @Param title query string false "Поиск по названию тура"
// @Param departure_city query string false "Город вылета"
// @Param trip_type query string false "Тип тура"
// @Param season query string false "Сезон"
// @Param route_city query string false "Город в маршруте"
// @Param active query bool false "Статус тура"
// @Param start_after query string false "Дата начала с (YYYY-MM-DD)"
// @Param end_before query string false "Дата окончания до (YYYY-MM-DD)"
// @Param limit query int false "Лимит (по умолчанию 20)"
// @Param offset query int false "Смещение"
// @Success 200 {object} models.TripPageResponse
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData
// @Router /trips/{id}/page [get]
func (h *TripPageHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	data, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, services.ErrTripNotFound) {
		h.log.Warnw("trip_page_not_found", "id", id)
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	}

	if err != nil {
		h.log.Errorw("trip_page_error", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось собрать данные страницы тура")
		return
	}

	helpers.JSON(w, http.StatusOK, data)
}

// ListAll
// @Summary All trips with full data
// @Description Возвращает список всех туров вместе с отелями, маршрутами и опциями
// @Tags Public — Trips
// @Produce json
// @Success 200 {array} models.TripPageResponse
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении туров"
// @Router /trips/full [get]
func (h *TripPageHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	f := parseTripFilter(r)
	trips, err := h.svc.ListAll(r.Context(), f)
	if err != nil {
		h.log.Errorw("trip_page_list_all_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить туры")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"items":   trips,
		"meta": map[string]interface{}{
			"limit":  f.Limit,
			"offset": f.Offset,
			"count":  len(trips),
		},
	})
}

// GetWithRelations
// @Summary Get trip with hotels and routes
// @Description Возвращает тур вместе с отелями и маршрутом
// @Tags Public — Trips
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} models.TripWithRelations
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData
// @Router /trips/{id}/relations [get]
func (h *TripPageHandler) GetWithRelations(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	data, err := h.svc.GetWithRelations(r.Context(), id)
	if errors.Is(err, services.ErrTripNotFound) {
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	}
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "Ошибка при получении тура")
		return
	}
	helpers.JSON(w, http.StatusOK, data)
}

// ListWithRelations
// @Summary List all trips with hotels and routes
// @Description Возвращает все туры вместе с отелями и маршрутами
// @Tags Public — Trips
// @Produce json
// @Param title query string false "Поиск по названию тура"
// @Param departure_city query string false "Город вылета"
// @Param trip_type query string false "Тип тура"
// @Param season query string false "Сезон"
// @Param route_city query string false "Город в маршруте"
// @Param active query bool false "Статус тура"
// @Param start_after query string false "Дата начала с (YYYY-MM-DD)"
// @Param end_before query string false "Дата окончания до (YYYY-MM-DD)"
// @Param limit query int false "Лимит (по умолчанию 20)"
// @Param offset query int false "Смещение"
// @Success 200 {array} models.TripWithRelations
// @Failure 500 {object} helpers.ErrorData
// @Router /trips/relations [get]
func (h *TripPageHandler) ListWithRelations(w http.ResponseWriter, r *http.Request) {
	f := parseTripFilter(r)
	data, err := h.svc.ListWithRelations(r.Context(), f)
	if err != nil {
		h.log.Errorw("trip_list_relations_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить туры")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"items":   data,
		"meta": map[string]interface{}{
			"limit":  f.Limit,
			"offset": f.Offset,
			"count":  len(data),
		},
	})
}

func parseTripFilter(r *http.Request) models.TripFilter {
	q := r.URL.Query()
	var f models.TripFilter

	f.Title = q.Get("title")
	f.DepartureCity = q.Get("departure_city")
	f.TripType = q.Get("trip_type")
	f.Season = q.Get("season")
	f.RouteCity = q.Get("route_city")

	if v := q.Get("active"); v != "" {
		val := v == "true" || v == "1"
		f.Active = &val
	}

	if v := q.Get("start_after"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.StartAfter = t
		}
	}
	if v := q.Get("end_before"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.EndBefore = t
		}
	}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Offset = n
		}
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
	return f
}
