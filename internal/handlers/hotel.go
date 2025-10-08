package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type HotelHandler struct {
	service *services.HotelService
	log     *zap.SugaredLogger
}

func NewHotelHandler(s *services.HotelService, log *zap.SugaredLogger) *HotelHandler {
	return &HotelHandler{service: s, log: log}
}

// Create
// @Summary Create hotel
// @Tags Admin — Hotels
// @Accept json
// @Produce json
// @Param hotel body models.HotelRequest true "Hotel"
// @Success 200 {object} models.HotelResponse
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 500 {object} helpers.ErrorData "Не удалось создать отель"
// @Router /admin/hotels [post]
func (h *HotelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.HotelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warnw("Некорректные данные при создании отеля", "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректные данные")
		return
	}

	hotel := models.Hotel{
		Name:     req.Name,
		City:     req.City,
		Stars:    req.Stars,
		Distance: req.Distance,
		Meals:    req.Meals,
		URLs:     req.URLs, // 👈 массив ссылок вместо photo_url
	}

	if req.DistanceText != nil {
		hotel.DistanceText = sql.NullString{String: *req.DistanceText, Valid: true}
	}
	if req.Guests != nil {
		hotel.Guests = sql.NullString{String: *req.Guests, Valid: true}
	}
	if req.Transfer != nil {
		hotel.Transfer = sql.NullString{String: *req.Transfer, Valid: true}
	}

	if err := h.service.Create(r.Context(), &hotel); err != nil {
		h.log.Errorw("Ошибка создания отеля", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось создать отель")
		return
	}

	h.log.Infow("Отель создан", "id", hotel.ID, "name", hotel.Name)
	helpers.JSON(w, http.StatusOK, toHotelResponse(hotel))
}

// List
// @Summary List hotels
// @Tags Admin — Hotels
// @Produce json
// @Success 200 {array} models.HotelResponse
// @Failure 500 {object} helpers.ErrorData "Не удалось получить список отелей"
// @Router /admin/hotels [get]
func (h *HotelHandler) List(w http.ResponseWriter, r *http.Request) {
	hotels, err := h.service.List(r.Context())
	if err != nil {
		h.log.Errorw("Ошибка получения списка отелей", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить список отелей")
		return
	}

	resp := make([]models.HotelResponse, 0, len(hotels))
	for _, hotel := range hotels {
		resp = append(resp, toHotelResponse(hotel))
	}

	h.log.Infow("Список отелей получен", "count", len(hotels))
	helpers.JSON(w, http.StatusOK, resp)
}

// Get
// @Summary Get hotel by ID
// @Tags Admin — Hotels
// @Produce json
// @Param id path int true "Hotel ID"
// @Success 200 {object} models.HotelResponse
// @Failure 404 {object} helpers.ErrorData "Отель не найден"
// @Router /admin/hotels/{id} [get]
func (h *HotelHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Warnw("Некорректный ID отеля", "id", chi.URLParam(r, "id"), "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	hotel, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.log.Warnw("Отель не найден", "id", id, "err", err)
		helpers.Error(w, http.StatusNotFound, "Отель не найден")
		return
	}

	h.log.Infow("Отель получен", "id", id)
	helpers.JSON(w, http.StatusOK, toHotelResponse(*hotel))
}

// Update
// @Summary Update hotel
// @Tags Admin — Hotels
// @Accept json
// @Produce json
// @Param id path int true "Hotel ID"
// @Param hotel body models.HotelRequest true "Hotel"
// @Success 200 {object} models.HotelResponse
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 500 {object} helpers.ErrorData "Не удалось обновить отель"
// @Router /admin/hotels/{id} [put]
func (h *HotelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Warnw("Некорректный ID отеля", "id", chi.URLParam(r, "id"), "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	var req models.HotelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warnw("Некорректные данные при обновлении отеля", "id", id, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректные данные")
		return
	}

	hotel := models.Hotel{
		ID:       id,
		Name:     req.Name,
		City:     req.City,
		Stars:    req.Stars,
		Distance: req.Distance,
		Meals:    req.Meals,
		URLs:     req.URLs, // 👈 массив ссылок
	}

	if req.DistanceText != nil {
		hotel.DistanceText = sql.NullString{String: *req.DistanceText, Valid: true}
	}
	if req.Guests != nil {
		hotel.Guests = sql.NullString{String: *req.Guests, Valid: true}
	}
	if req.Transfer != nil {
		hotel.Transfer = sql.NullString{String: *req.Transfer, Valid: true}
	}

	if err := h.service.Update(r.Context(), &hotel); err != nil {
		h.log.Errorw("Ошибка обновления отеля", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить отель")
		return
	}

	h.log.Infow("Отель обновлён", "id", id)
	helpers.JSON(w, http.StatusOK, toHotelResponse(hotel))
}

// Delete
// @Summary Delete hotel
// @Tags Admin — Hotels
// @Param id path int true "Hotel ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData "Некорректный ID"
// @Failure 500 {object} helpers.ErrorData "Не удалось удалить отель"
// @Router /admin/hotels/{id} [delete]
func (h *HotelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Warnw("Некорректный ID отеля", "id", chi.URLParam(r, "id"), "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.log.Errorw("Ошибка удаления отеля", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить отель")
		return
	}

	h.log.Infow("Отель удалён", "id", id)
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// AttachHotelToTrip
// @Summary Attach hotel to trip
// @Tags Admin — Trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param body body models.TripHotel true "Hotel ID and Nights"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 500 {object} helpers.ErrorData "Не удалось привязать отель"
// @Router /admin/trips/{id}/hotels [post]
func (h *HotelHandler) AttachHotelToTrip(w http.ResponseWriter, r *http.Request) {
	tripID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		h.log.Warnw("Некорректный ID тура", "id", chi.URLParam(r, "id"), "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID тура")
		return
	}

	var th models.TripHotel
	if err := json.NewDecoder(r.Body).Decode(&th); err != nil {
		h.log.Warnw("Некорректные данные при привязке отеля к туру", "trip_id", tripID, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректные данные")
		return
	}
	th.TripID = tripID

	if th.Nights <= 0 {
		th.Nights = 1
	}

	if err := h.service.Attach(r.Context(), &th); err != nil {
		h.log.Errorw("Ошибка привязки отеля к туру", "trip_id", tripID, "hotel_id", th.HotelID, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось привязать отель")
		return
	}

	h.log.Infow("Отель привязан к туру", "trip_id", tripID, "hotel_id", th.HotelID, "nights", th.Nights)
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "attached"})
}

func toHotelResponse(h models.Hotel) models.HotelResponse {
	var distanceText, guests, transfer *string
	if h.DistanceText.Valid {
		distanceText = &h.DistanceText.String
	}
	if h.Guests.Valid {
		guests = &h.Guests.String
	}
	if h.Transfer.Valid {
		transfer = &h.Transfer.String
	}

	return models.HotelResponse{
		ID:           h.ID,
		Name:         h.Name,
		City:         h.City,
		Stars:        h.Stars,
		Distance:     h.Distance,
		DistanceText: distanceText,
		Meals:        h.Meals,
		Guests:       guests,
		URLs:         h.URLs, // 👈 массив ссылок
		Transfer:     getOrDefault(transfer, "не указано"),
		Nights:       h.Nights,
		CreatedAt:    h.CreatedAt,
		UpdatedAt:    h.UpdatedAt,
	}
}

func getOrDefault(s *string, def string) string {
	if s != nil {
		return *s
	}
	return def
}
