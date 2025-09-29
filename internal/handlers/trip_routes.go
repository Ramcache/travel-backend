package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
)

type TripRouteHandler struct {
	svc *services.TripRouteService
	log *zap.SugaredLogger
}

func NewTripRouteHandler(svc *services.TripRouteService, log *zap.SugaredLogger) *TripRouteHandler {
	return &TripRouteHandler{svc: svc, log: log}
}

// Create
// @Summary Создать маршрут тура
// @Description Добавляет новый маршрут к туру
// @Tags admin, trips, routes
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param body body models.TripRouteRequest true "Маршрут"
// @Success 201 {object} models.TripRoute
// @Failure 400 {object} helpers.ErrorData "Некорректный JSON"
// @Failure 500 {object} helpers.ErrorData "Ошибка при добавлении маршрута"
// @Router /admin/trips/{id}/routes [post]
func (h *TripRouteHandler) Create(w http.ResponseWriter, r *http.Request) {
	tripID, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var req models.TripRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	route, err := h.svc.Create(r.Context(), tripID, req)
	if err != nil {
		h.log.Errorw("trip_route_create_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось добавить маршрут")
		return
	}

	helpers.JSON(w, http.StatusCreated, route)
}

// Update
// @Summary Обновить маршрут тура
// @Description Обновляет существующий маршрут
// @Tags admin, trips, routes
// @Accept json
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param route_id path int true "Route ID"
// @Param body body models.TripRouteRequest true "Маршрут"
// @Success 200 {object} models.TripRoute
// @Failure 400 {object} helpers.ErrorData "Некорректный JSON"
// @Failure 500 {object} helpers.ErrorData "Ошибка при обновлении маршрута"
// @Router /admin/trips/{trip_id}/routes/{route_id} [put]
func (h *TripRouteHandler) Update(w http.ResponseWriter, r *http.Request) {
	routeID, _ := strconv.Atoi(chi.URLParam(r, "route_id"))

	var req models.TripRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	route, err := h.svc.Update(r.Context(), routeID, req)
	if err != nil {
		h.log.Errorw("trip_route_update_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить маршрут")
		return
	}

	helpers.JSON(w, http.StatusOK, route)
}

// Delete
// @Summary Удалить маршрут тура
// @Description Удаляет маршрут тура по ID
// @Tags admin, trips, routes
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param route_id path int true "Route ID"
// @Success 200 {object} map[string]string "Маршрут удалён"
// @Failure 500 {object} helpers.ErrorData "Ошибка при удалении маршрута"
// @Router /admin/trips/{trip_id}/routes/{route_id} [delete]
func (h *TripRouteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	routeID, _ := strconv.Atoi(chi.URLParam(r, "route_id"))

	if err := h.svc.Delete(r.Context(), routeID); err != nil {
		h.log.Errorw("trip_route_delete_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить маршрут")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"message": "Маршрут удалён"})
}

// List
// @Summary Список маршрутов тура
// @Description Возвращает список маршрутов для тура
// @Tags trips, routes
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {array} models.TripRoute
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении маршрутов"
// @Router /trips/{id}/routes [get]
func (h *TripRouteHandler) List(w http.ResponseWriter, r *http.Request) {
	tripID, _ := strconv.Atoi(chi.URLParam(r, "id"))

	routes, err := h.svc.ListByTrip(r.Context(), tripID)
	if err != nil {
		h.log.Errorw("trip_routes_list_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить маршрут")
		return
	}

	helpers.JSON(w, http.StatusOK, routes)
}
