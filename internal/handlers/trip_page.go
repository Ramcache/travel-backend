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
