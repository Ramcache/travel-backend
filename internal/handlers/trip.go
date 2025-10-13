package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
	"github.com/Ramcache/travel-backend/internal/services"
)

type TripHandler struct {
	service      services.TripServiceI
	orderService *services.OrderService
	hotelService *services.HotelService
	log          *zap.SugaredLogger
}

func NewTripHandler(service services.TripServiceI, orderService *services.OrderService, hotelService *services.HotelService, log *zap.SugaredLogger) *TripHandler {
	return &TripHandler{service: service, orderService: orderService, hotelService: hotelService, log: log}
}

// List
// @Summary List trips
// @Description Публичный поиск туров с фильтрацией и пагинацией
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
// @Success 200 {object} map[string]interface{} "success + items + meta"
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении списка туров"
// @Router /trips [get]
func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
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

	trips, err := h.service.List(r.Context(), f)
	if err != nil {
		h.log.Errorw("Ошибка получения списка туров", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить список туров")
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

// Get
// @Summary Get trip by id
// @Description Публичный просмотр тура
// @Tags Public — Trips
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} models.Trip
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении тура"
// @Router /trips/{id} [get]
func (h *TripHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	trip, err := h.service.Get(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrTripNotFound):
		h.log.Warnw("Тур не найден", "id", id)
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	case err != nil:
		h.log.Errorw("Ошибка получения тура", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить тур")
		return
	}
	go func(id int) {
		if err := h.service.IncrementViews(context.Background(), id); err != nil {
			h.log.Errorw("increment_views_failed", "id", id, "err", err)
		}
	}(id)

	h.log.Infow("Тур успешно получен", "id", id)
	helpers.JSON(w, http.StatusOK, trip)
}

// Create
// @Summary Create trip (admin)
// @Description Создание нового тура (только админ)
// @Tags Admin — Trips
// @Security Bearer
// @Accept json
// @Produce json
// @Param data body models.CreateTripRequest true "Trip data"
// @Success 200 {object} models.Trip
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/trips [post]
func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Некорректный JSON при создании тура", "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	trip, err := h.service.Create(r.Context(), req)
	switch {
	case helpers.IsInvalidInput(err):
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка создания тура", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось создать тур")
		return
	}

	h.log.Infow("Тур успешно создан", "id", trip.ID)
	helpers.JSON(w, http.StatusCreated, trip)
}

// Update
// @Summary Update trip (admin)
// @Tags Admin — Trips
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param data body models.UpdateTripRequest true "Trip update"
// @Success 200 {object} models.Trip
// @Failure 400 {object} helpers.ErrorData
// @Failure 404 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/trips/{id} [put]
func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	trip, err := h.service.Update(r.Context(), id, req)
	switch {
	case errors.Is(err, services.ErrTripNotFound):
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	case helpers.IsInvalidInput(err):
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка обновления тура", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить тур")
		return
	}

	helpers.JSON(w, http.StatusOK, trip)
}

// Delete
// @Summary Delete trip (admin)
// @Tags Admin — Trips
// @Security Bearer
// @Param id path int true "Trip ID"
// @Success 204
// @Failure 404 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/trips/{id} [delete]
func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	err := h.service.Delete(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrTripNotFound):
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	case err != nil:
		h.log.Errorw("Ошибка удаления тура", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить тур")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Countdown
// @Summary Get booking countdown
// @Description Получить обратный отсчёт до конца бронирования
// @Tags Public — Trips
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} map[string]int
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении тура"
// @Router /trips/{id}/countdown [get]
func (h *TripHandler) Countdown(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	trip, err := h.service.Get(r.Context(), id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	}

	now := time.Now()
	diff := trip.BookingDeadline.Sub(now)
	if diff < 0 {
		helpers.JSON(w, http.StatusOK, map[string]int{
			"days": 0, "hours": 0, "minutes": 0, "seconds": 0,
		})
		return
	}

	days := int(diff.Hours()) / 24
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60

	helpers.JSON(w, http.StatusOK, map[string]int{
		"days": days, "hours": hours, "minutes": minutes, "seconds": seconds,
	})
}

// GetMain
// @Summary Get main trip with countdown
// @Description Получить главный тур для главной страницы (только название и обратный отсчёт)
// @Tags Public — Trips
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} helpers.ErrorData "Главный тур не найден"
// @Router /trips/main [get]
func (h *TripHandler) GetMain(w http.ResponseWriter, r *http.Request) {
	trip, err := h.service.GetMain(r.Context())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			helpers.Error(w, http.StatusNotFound, "Главный тур не найден")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "Ошибка при получении главного тура")
		return
	}

	var days, hours, minutes, seconds int
	if trip.BookingDeadline != nil {
		diff := trip.BookingDeadline.Sub(time.Now())
		if diff > 0 {
			days = int(diff.Hours()) / 24
			hours = int(diff.Hours()) % 24
			minutes = int(diff.Minutes()) % 60
			seconds = int(diff.Seconds()) % 60
		}
	}

	helpers.JSON(w, http.StatusOK, map[string]interface{}{
		"title": trip.Title,
		"countdown": map[string]int{
			"days": days, "hours": hours, "minutes": minutes, "seconds": seconds,
		},
	})
}

// Popular
// @Summary Get popular trips
// @Tags Public — Trips
// @Produce json
// @Param limit query int false "Количество туров (по умолчанию 5)"
// @Success 200 {array} models.Trip
// @Failure 500 {object} helpers.ErrorData "Не удалось получить популярные туры"
// @Router /trips/popular [get]
func (h *TripHandler) Popular(w http.ResponseWriter, r *http.Request) {
	limit := 5
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}
	trips, err := h.service.Popular(r.Context(), limit)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить популярные туры")
		return
	}
	helpers.JSON(w, http.StatusOK, trips)
}

// Buy
// @Summary Buy trip
// @Description Отправка заявки на покупку тура в Telegram
// @Tags Public — Trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param data body models.BuyRequest true "Данные покупателя"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при покупке тура"
// @Router /trips/{id}/buy [post]
func (h *TripHandler) Buy(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.BuyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}
	if err := h.service.Buy(r.Context(), id, req); err != nil {
		if errors.Is(err, services.ErrTripNotFound) {
			helpers.Error(w, http.StatusNotFound, "Тур не найден")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "Ошибка при покупке тура")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// BuyWithoutTrip
// @Summary Buy trip
// @Description Отправка заявки на покупку тура в Telegram
// @Tags Public — Trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param data body models.BuyRequest true "Данные покупателя"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при покупке тура"
// @Router /trips/{id}/buy [post]
func (h *TripHandler) BuyWithoutTrip(w http.ResponseWriter, r *http.Request) {
	var req models.BuyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}
	if err := h.service.BuyWithoutTrip(r.Context(), req); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "Ошибка при покупке без тура")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// CreateTour — создаёт тур, отель и маршрут за один запрос
// @Summary Create Tour with Hotel and Route
// @Description Админская ручка: создаёт тур, отель и маршрут одним запросом
// @Tags Admin — Trips
// @Accept json
// @Produce json
// @Param data body models.CreateTourRequest true "Tour + Hotel + Route"
// @Success 201 {object} models.CreateTourResponse
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/tours [post]
func (h *TripHandler) CreateTour(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTourRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	ctx := r.Context()

	// === Создаём тур ===
	trip, err := h.service.Create(ctx, req.Trip)
	if err != nil {
		h.log.Errorw("create_tour_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Ошибка создания тура")
		return
	}

	var hotels []models.HotelResponse

	// === Обрабатываем отели ===
	for _, hreq := range req.Hotels {
		// определяем количество ночей
		nights := hreq.Nights
		if nights == 0 {
			nights = 1
		}

		// 1️⃣ Если передан hotel_id — прикрепляем существующий отель
		if hreq.HotelID > 0 {
			th := &models.TripHotel{
				TripID:  trip.ID,
				HotelID: hreq.HotelID,
				Nights:  nights,
			}
			if err := h.hotelService.Attach(ctx, th); err != nil {
				h.log.Errorw("attach_existing_hotel_failed",
					"trip_id", trip.ID,
					"hotel_id", hreq.HotelID,
					"err", err,
				)
				helpers.Error(w, http.StatusInternalServerError, "Ошибка привязки существующего отеля к туру")
				return
			}

			// ✅ Подтягиваем отель из базы
			hotel, err := h.hotelService.GetByID(ctx, hreq.HotelID)
			if err != nil {
				h.log.Warnw("hotel_not_found_after_attach", "hotel_id", hreq.HotelID, "err", err)
				continue
			}

			hotel.Nights = nights
			hotels = append(hotels, toHotelResponse(*hotel))
			continue
		}

		// 2️⃣ Если hotel_id нет — создаём новый отель
		hotel := models.Hotel{
			Name:     hreq.Name,
			City:     hreq.City,
			Stars:    hreq.Stars,
			Distance: hreq.Distance,
			Meals:    hreq.Meals,
			URLs:     hreq.URLs,
		}

		if hreq.DistanceText != nil {
			hotel.DistanceText = sql.NullString{String: *hreq.DistanceText, Valid: true}
		}
		if hreq.Guests != nil {
			hotel.Guests = sql.NullString{String: *hreq.Guests, Valid: true}
		}
		if hreq.Transfer != nil {
			hotel.Transfer = sql.NullString{String: *hreq.Transfer, Valid: true}
		}

		if err := h.service.CreateHotel(ctx, &hotel); err != nil {
			h.log.Errorw("create_hotel_failed", "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка создания отеля")
			return
		}

		th := &models.TripHotel{
			TripID:  trip.ID,
			HotelID: hotel.ID,
			Nights:  nights,
		}
		if err := h.hotelService.Attach(ctx, th); err != nil {
			h.log.Errorw("attach_new_hotel_failed",
				"trip_id", trip.ID,
				"hotel_id", hotel.ID,
				"err", err,
			)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка привязки отеля к туру")
			return
		}

		hotel.Nights = nights

		hotels = append(hotels, toHotelResponse(hotel))

	}

	// === Обработка маршрутов ===
	var routes []models.TripRoute

	// новый формат (routes)
	if len(req.Routes) > 0 {
		for _, rreq := range req.Routes {
			rt, err := h.service.CreateRoute(ctx, trip.ID, rreq)
			if err != nil {
				h.log.Errorw("create_route_failed", "trip_id", trip.ID, "err", err)
				helpers.Error(w, http.StatusInternalServerError, "Ошибка создания маршрута")
				return
			}
			routes = append(routes, *rt)
		}
	} else {
		// старый формат (route_cities)
		routeReqs := models.ConvertCitiesToRoutes(req.RouteCities)
		for _, rreq := range routeReqs {
			rt, err := h.service.CreateRoute(ctx, trip.ID, rreq)
			if err != nil {
				h.log.Errorw("create_route_failed", "trip_id", trip.ID, "err", err)
				helpers.Error(w, http.StatusInternalServerError, "Ошибка создания маршрута")
				return
			}
			routes = append(routes, *rt)
		}
	}

	routeResp := models.ConvertRoutesToCities(routes)

	// === Успешный ответ ===
	helpers.JSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"trip":    trip,
		"hotels":  hotels,
		"routes":  routeResp,
	})
}

// GetFull
// @Summary Получить тур с отелями и маршрутами
// @Tags Admin — Trips
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} models.TripFullResponse
// @Failure 404 {object} helpers.ErrorData
// @Router /admin/trips/{id}/full [get]
func (h *TripHandler) GetFull(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID тура")
		return
	}
	resp, err := h.service.GetFull(r.Context(), id)
	if err != nil {
		h.log.Errorw("trip_full_get_failed", "id", id, "err", err)
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	}
	helpers.JSON(w, http.StatusOK, resp)
}

// UpdateTour
// @Summary Обновить тур, отели и маршруты одной кнопкой
// @Tags Admin — Trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param body body models.UpdateTourRequest true "Trip with hotels and routes"
// @Success 200 {object} models.Trip
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/trips/{id}/full [put]
// UpdateTour обновляет тур вместе с отелями и маршрутами (full update)
func (h *TripHandler) UpdateTour(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tripID, err := strconv.Atoi(idParam)
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID тура")
		return
	}

	var req models.UpdateTourRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	ctx := r.Context()

	// 1️⃣ Обновляем сам тур
	trip, err := h.service.Update(ctx, tripID, req.Trip)
	if err != nil {
		h.log.Errorw("update_tour_failed", "trip_id", tripID, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Ошибка обновления тура")
		return
	}

	// 2️⃣ Обновляем / привязываем отели (полная замена)
	var hotels []models.HotelResponse

	if req.Hotels != nil {
		affected, err := h.hotelService.ClearByTrip(ctx, tripID)
		if err != nil {
			h.log.Errorw("clear_trip_hotels_failed", "trip_id", tripID, "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка очистки старых отелей тура")
			return
		}
		h.log.Infow("trip_hotels_cleared", "trip_id", tripID, "deleted_rows", affected)

		for _, hreq := range req.Hotels {
			nights := hreq.Nights
			if nights == 0 {
				nights = 1
			}

			if hreq.HotelID > 0 {
				// если отель уже существует — привязываем
				th := &models.TripHotel{
					TripID:  tripID,
					HotelID: hreq.HotelID,
					Nights:  nights,
				}
				if err := h.hotelService.Attach(ctx, th); err != nil {
					h.log.Errorw("attach_hotel_failed", "trip_id", tripID, "hotel_id", hreq.HotelID, "err", err)
					helpers.Error(w, http.StatusInternalServerError, "Ошибка привязки отеля")
					return
				}

				hotel, err := h.hotelService.GetByID(ctx, hreq.HotelID)
				if err != nil {
					h.log.Warnw("hotel_not_found_for_response", "hotel_id", hreq.HotelID, "err", err)
				} else {
					hotel.Nights = nights
					hotels = append(hotels, toHotelResponse(*hotel))
				}
			} else {
				// иначе создаём новый отель и привязываем его
				hotel := models.Hotel{
					Name:     hreq.Name,
					City:     hreq.City,
					Stars:    hreq.Stars,
					Distance: hreq.Distance,
					Meals:    hreq.Meals,
					URLs:     hreq.URLs,
				}
				if err := h.service.CreateHotel(ctx, &hotel); err != nil {
					h.log.Errorw("create_hotel_failed", "err", err)
					helpers.Error(w, http.StatusInternalServerError, "Ошибка создания отеля")
					return
				}

				th := &models.TripHotel{
					TripID:  tripID,
					HotelID: hotel.ID,
					Nights:  nights,
				}
				if err := h.hotelService.Attach(ctx, th); err != nil {
					h.log.Errorw("attach_new_hotel_failed", "trip_id", tripID, "hotel_id", hotel.ID, "err", err)
					helpers.Error(w, http.StatusInternalServerError, "Ошибка привязки отеля")
					return
				}

				hotel.Nights = nights
				hotels = append(hotels, toHotelResponse(hotel))
			}
		}
	}
	// 3️⃣ обновляем маршруты (включая legacy route_cities)
	var routes []models.TripRoute

	// Определяем, какие данные пришли — новые или старые
	hasNewRoutes := req.Routes != nil
	hasLegacyRoutes := req.Routes == nil && len(req.RouteCities) > 0

	if hasNewRoutes || hasLegacyRoutes {
		// 💥 Очистка старых маршрутов перед обновлением
		affected, err := h.service.ClearRoutesByTrip(ctx, tripID)
		if err != nil {
			h.log.Errorw("clear_trip_routes_failed", "trip_id", tripID, "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка очистки старых маршрутов тура")
			return
		}
		h.log.Infow("trip_routes_cleared", "trip_id", tripID, "deleted_rows", affected)

		if hasNewRoutes && len(req.Routes) > 0 {
			for _, rreq := range req.Routes {
				rt, err := h.service.CreateRoute(ctx, tripID, rreq)
				if err != nil {
					h.log.Errorw("update_route_failed", "trip_id", tripID, "err", err)
					helpers.Error(w, http.StatusInternalServerError, "Ошибка обновления маршрута")
					return
				}
				routes = append(routes, *rt)
			}
		} else if hasLegacyRoutes {
			// Поддержка старого поля route_cities
			routeReqs := models.ConvertCitiesToRoutes(req.RouteCities)
			for _, rreq := range routeReqs {
				rt, err := h.service.CreateRoute(ctx, tripID, rreq)
				if err != nil {
					h.log.Errorw("update_route_failed", "trip_id", tripID, "err", err)
					helpers.Error(w, http.StatusInternalServerError, "Ошибка обновления маршрута")
					return
				}
				routes = append(routes, *rt)
			}
		}
	}

	routeResp := models.ConvertRoutesToCities(routes)

	helpers.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"trip":    trip,
		"hotels":  hotels,
		"routes":  routeResp,
	})
}
