package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/go-chi/chi/v5"
)

type HotelHandler struct {
	service *services.HotelService
}

func NewHotelHandler(s *services.HotelService) *HotelHandler {
	return &HotelHandler{service: s}
}

// Create
// @Summary Create hotel
// @Tags hotels
// @Accept json
// @Produce json
// @Param hotel body models.Hotel true "Hotel"
// @Success 200 {object} models.Hotel
// @Failure 400 {object} helpers.ErrorData
// @Router /admin/hotels [post]
func (h *HotelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var hotel models.Hotel
	if err := json.NewDecoder(r.Body).Decode(&hotel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.service.Create(r.Context(), &hotel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(hotel)
}

// List
// @Summary List hotels
// @Tags hotels
// @Produce json
// @Success 200 {array} models.Hotel
// @Router /admin/hotels [get]
func (h *HotelHandler) List(w http.ResponseWriter, r *http.Request) {
	hotels, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(hotels)
}

// Get
// @Summary Get hotel by ID
// @Tags hotels
// @Produce json
// @Param id path int true "Hotel ID"
// @Success 200 {object} models.Hotel
// @Failure 404 {object} helpers.ErrorData
// @Router /admin/hotels/{id} [get]
func (h *HotelHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	hotel, err := h.service.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(hotel)
}

// Update
// @Summary Update hotel
// @Tags hotels
// @Accept json
// @Produce json
// @Param id path int true "Hotel ID"
// @Param hotel body models.Hotel true "Hotel"
// @Success 200 {object} models.Hotel
// @Failure 400 {object} helpers.ErrorData
// @Router /admin/hotels/{id} [put]
func (h *HotelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var hotel models.Hotel
	if err := json.NewDecoder(r.Body).Decode(&hotel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hotel.ID = id
	if err := h.service.Update(r.Context(), &hotel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(hotel)
}

// Delete
// @Summary Delete hotel
// @Tags hotels
// @Param id path int true "Hotel ID"
// @Success 204 {string} string "No Content"
// @Router /admin/hotels/{id} [delete]
func (h *HotelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AttachHotelToTrip
// @Summary Attach hotel to trip
// @Tags trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param body body models.TripHotel true "Hotel ID and Nights"
// @Success 200 {string} string "attached"
// @Router /admin/trips/{id}/hotels [post]
func (h *HotelHandler) AttachHotelToTrip(w http.ResponseWriter, r *http.Request) {
	tripID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var th models.TripHotel
	if err := json.NewDecoder(r.Body).Decode(&th); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	th.TripID = tripID
	if err := h.service.Attach(r.Context(), &th); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`{"status":"attached"}`))
}
