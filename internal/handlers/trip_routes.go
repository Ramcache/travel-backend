package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
)

type TripRouteHandler struct {
	svc      *services.TripRouteService
	log      *zap.SugaredLogger
	validate *validator.Validate
}

func NewTripRouteHandler(svc *services.TripRouteService, log *zap.SugaredLogger) *TripRouteHandler {
	return &TripRouteHandler{svc: svc, log: log, validate: validator.New()}
}

// ListUI
// @Summary UI-маршрут тура (для плашки)
// @Tags trips, routes
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} models.TripRouteUIResponse
// @Failure 500 {object} helpers.ErrorData
// @Router /trips/{id}/routes/ui [get]
func (h *TripRouteHandler) ListUI(w http.ResponseWriter, r *http.Request) {
	tripID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	resp, err := h.svc.GetUIRoute(r.Context(), tripID)
	if err != nil {
		h.log.Errorw("trip_routes_ui_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить маршрут")
		return
	}
	helpers.JSON(w, http.StatusOK, resp)
}

// --- Совместимый список ---
func (h *TripRouteHandler) List(w http.ResponseWriter, r *http.Request) {
	tripID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	resp, err := h.svc.GetRouteResponse(r.Context(), tripID)
	if err != nil {
		h.log.Errorw("trip_routes_list_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить маршрут")
		return
	}
	helpers.JSON(w, http.StatusOK, resp)
}

// CreateBatch
// @Summary Создать несколько маршрутов тура
// @Tags admin, trips, routes
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param body body models.TripRouteBatchRequest true "Маршруты"
// @Success 201 {array} models.TripRoute
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/trips/{id}/routes/batch [post]
func (h *TripRouteHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	tripID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var batch models.TripRouteBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	for _, rt := range batch.Routes {
		if err := h.validate.Struct(rt); err != nil {
			helpers.Error(w, http.StatusBadRequest, "Неверные данные: "+err.Error())
			return
		}
	}
	routes, err := h.svc.CreateBatch(r.Context(), tripID, batch.Routes)
	if err != nil {
		h.log.Errorw("trip_route_batch_create_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось добавить маршруты")
		return
	}
	helpers.JSON(w, http.StatusCreated, routes)
}

// Update
// @Summary Обновить маршрут тура
// @Tags admin, trips, routes
// @Accept json
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param route_id path int true "Route ID"
// @Param body body models.TripRouteRequest true "Маршрут"
// @Success 200 {object} models.TripRoute
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/trips/{trip_id}/routes/{route_id} [put]
func (h *TripRouteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "route_id"))
	var req models.TripRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	route, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.log.Errorw("trip_route_update_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить маршрут")
		return
	}
	helpers.JSON(w, http.StatusOK, route)
}

// Delete
// @Summary Удалить маршрут тура
// @Tags admin, trips, routes
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param route_id path int true "Route ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/trips/{trip_id}/routes/{route_id} [delete]
func (h *TripRouteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "route_id"))
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.log.Errorw("trip_route_delete_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить маршрут")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"message": "Маршрут удалён"})
}
