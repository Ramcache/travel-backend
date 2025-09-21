package handlers

import (
	"encoding/json"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"

	"github.com/Ramcache/travel-backend/internal/helpers"
)

type TripHandler struct {
	service *services.TripService
}

func NewTripHandler(service *services.TripService) *TripHandler {
	return &TripHandler{service: service}
}

// List
// Public search
// @Summary List trips
// @Tags trips
// @Produce json
// @Param departure_city query string false "Город вылета"
// @Param trip_type query string false "Тип тура"
// @Param season query string false "Сезон"
// @Success 200 {array} models.Trip
// @Router /trips [get]
func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("departure_city")
	ttype := r.URL.Query().Get("trip_type")
	season := r.URL.Query().Get("season")

	trips, err := h.service.List(r.Context(), city, ttype, season)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.JSON(w, http.StatusOK, trips)
}

// Get
// @Summary Get trip by id
// @Tags trips
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} models.Trip
// @Router /trips/{id} [get]
func (h *TripHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	trip, err := h.service.Get(r.Context(), id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "trip not found")
		return
	}
	helpers.JSON(w, http.StatusOK, trip)
}

// Create
// @Summary Create trip (admin)
// @Tags trips
// @Security Bearer
// @Accept json
// @Produce json
// @Param data body models.CreateTripRequest true "Trip data"
// @Success 200 {object} models.Trip
// @Router /admin/trips [post]
func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}
	trip, err := h.service.Create(r.Context(), req)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.JSON(w, http.StatusOK, trip)
}

// Update
// @Summary Update trip (admin)
// @Tags trips
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param data body models.UpdateTripRequest true "Trip update"
// @Success 200 {object} models.Trip
// @Router /admin/trips/{id} [put]
func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}
	trip, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.JSON(w, http.StatusOK, trip)
}

// Delete
// @Summary Delete trip (admin)
// @Tags trips
// @Security Bearer
// @Param id path int true "Trip ID"
// @Success 204 "No Content"
// @Router /admin/trips/{id} [delete]
func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.service.Delete(r.Context(), id); err != nil {
		helpers.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// helper
func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
