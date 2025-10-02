package handlers

import (
	"errors"
	"net/http"
	"strconv"

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
// @Tags trips
// @Produce json
// @Param id path int true "Trip ID"
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
// @Tags trips
// @Produce json
// @Success 200 {array} models.TripPageResponse
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении туров"
// @Router /trips/full [get]
func (h *TripPageHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	trips, err := h.svc.ListAll(r.Context())
	if err != nil {
		h.log.Errorw("trip_page_list_all_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить туры")
		return
	}
	helpers.JSON(w, http.StatusOK, trips)
}

// GetWithRelations
// @Summary Get trip with hotels and routes
// @Description Возвращает тур вместе с отелями и маршрутом
// @Tags trips
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
// @Tags trips
// @Produce json
// @Success 200 {array} models.TripWithRelations
// @Failure 500 {object} helpers.ErrorData
// @Router /trips/relations [get]
func (h *TripPageHandler) ListWithRelations(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListWithRelations(r.Context())
	if err != nil {
		h.log.Errorw("trip_list_relations_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить туры")
		return
	}
	helpers.JSON(w, http.StatusOK, data)
}
