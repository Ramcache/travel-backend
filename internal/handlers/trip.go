package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Ramcache/travel-backend/internal/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
)

type TripHandler struct {
	service      services.TripServiceI
	orderService *services.OrderService
	log          *zap.SugaredLogger
}

func NewTripHandler(service services.TripServiceI, orderService *services.OrderService, log *zap.SugaredLogger) *TripHandler {
	return &TripHandler{service: service, orderService: orderService, log: log}
}

// List
// @Summary List trips
// @Description Публичный поиск туров с фильтрацией
// @Tags trips
// @Produce json
// @Param departure_city query string false "Город вылета"
// @Param trip_type query string false "Тип тура (пляжный, экскурсионный, семейный)"
// @Param season query string false "Сезон (например: 2025 или лето 2025)"
// @Success 200 {array} models.Trip
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении списка туров"
// @Router /trips [get]
func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("departure_city")
	ttype := r.URL.Query().Get("trip_type")
	season := r.URL.Query().Get("season")

	trips, err := h.service.List(r.Context(), city, ttype, season)
	if err != nil {
		h.log.Errorw("Ошибка получения списка туров", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить список туров")
		return
	}

	h.log.Infow("Список туров успешно получен", "count", len(trips))
	helpers.JSON(w, http.StatusOK, trips)
}

// Get
// @Summary Get trip by id
// @Description Публичный просмотр тура
// @Tags trips
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
// @Tags trips
// @Security Bearer
// @Accept json
// @Produce json
// @Param data body models.CreateTripRequest true "Trip data"
// @Success 200 {object} models.Trip
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 500 {object} helpers.ErrorData "Ошибка при создании тура"
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
		h.log.Warnw("Ошибка валидации при создании тура", "err", err)
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка создания тура", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось создать тур")
		return
	}

	h.log.Infow("Тур успешно создан", "id", trip.ID)
	helpers.JSON(w, http.StatusOK, trip)
}

// Update
// @Summary Update trip (admin)
// @Description Обновление данных тура (только админ)
// @Tags trips
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param data body models.UpdateTripRequest true "Trip update"
// @Success 200 {object} models.Trip
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при обновлении тура"
// @Router /admin/trips/{id} [put]
func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorw("Некорректный JSON при обновлении тура", "id", id, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	trip, err := h.service.Update(r.Context(), id, req)
	switch {
	case errors.Is(err, services.ErrTripNotFound):
		h.log.Warnw("Тур не найден для обновления", "id", id)
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	case helpers.IsInvalidInput(err):
		h.log.Warnw("Ошибка валидации при обновлении тура", "id", id, "err", err)
		helpers.Error(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		h.log.Errorw("Ошибка обновления тура", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить тур")
		return
	}

	h.log.Infow("Тур успешно обновлён", "id", id)
	helpers.JSON(w, http.StatusOK, trip)
}

// Delete
// @Summary Delete trip (admin)
// @Description Удаление тура (только админ)
// @Tags trips
// @Security Bearer
// @Param id path int true "Trip ID"
// @Success 204 "No Content"
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при удалении тура"
// @Router /admin/trips/{id} [delete]
func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	err := h.service.Delete(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrTripNotFound):
		h.log.Warnw("Тур не найден для удаления", "id", id)
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	case err != nil:
		h.log.Errorw("Ошибка удаления тура", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить тур")
		return
	}

	h.log.Infow("Тур успешно удалён", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

// Countdown
// @Summary Get booking countdown
// @Description Получить обратный отсчёт до конца бронирования
// @Tags trips
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} map[string]int
// @Failure 404 {object} helpers.ErrorData "Тур не найден"
// @Failure 500 {object} helpers.ErrorData "Ошибка при получении тура"
// @Router /trips/{id}/countdown [get]
func (h *TripHandler) Countdown(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	trip, err := h.service.Get(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrTripNotFound):
		h.log.Warnw("Тур не найден при countdown", "id", id)
		helpers.Error(w, http.StatusNotFound, "Тур не найден")
		return
	case err != nil:
		h.log.Errorw("Ошибка получения тура при countdown", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить тур")
		return
	}

	now := time.Now()
	diff := trip.BookingDeadline.Sub(now)
	if diff < 0 {
		h.log.Infow("Срок бронирования истёк", "id", id)
		helpers.JSON(w, http.StatusOK, map[string]int{
			"days": 0, "hours": 0, "minutes": 0, "seconds": 0,
		})
		return
	}

	days := int(diff.Hours()) / 24
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60

	h.log.Infow("Обратный отсчёт успешно рассчитан",
		"id", id,
		"days", days,
		"hours", hours,
		"minutes", minutes,
		"seconds", seconds,
	)

	helpers.JSON(w, http.StatusOK, map[string]int{
		"days":    days,
		"hours":   hours,
		"minutes": minutes,
		"seconds": seconds,
	})
}

// GetMain
// @Summary Get main trip with countdown
// @Description Получить главный тур для главной страницы (только название и обратный отсчёт)
// @Tags trips
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

	// считаем countdown
	var days, hours, minutes, seconds int
	if trip.BookingDeadline != nil {
		now := time.Now()
		diff := trip.BookingDeadline.Sub(now)
		if diff > 0 {
			days = int(diff.Hours()) / 24
			hours = int(diff.Hours()) % 24
			minutes = int(diff.Minutes()) % 60
			seconds = int(diff.Seconds()) % 60
		}
	}

	response := map[string]interface{}{
		"title": trip.Title,
		"countdown": map[string]int{
			"days":    days,
			"hours":   hours,
			"minutes": minutes,
			"seconds": seconds,
		},
	}

	h.log.Infow("Главный тур с countdown возвращён", "id", trip.ID, "title", trip.Title)
	helpers.JSON(w, http.StatusOK, response)
}

// Popular
// @Summary Get popular trips
// @Tags trips
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
		h.log.Errorw("Ошибка получения популярных туров", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить популярные туры")
		return
	}
	helpers.JSON(w, http.StatusOK, trips)
}

// Buy
// @Summary Buy trip
// @Description Отправка заявки на покупку тура в Telegram
// @Tags trips
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
		h.log.Errorw("Некорректный JSON при покупке тура", "id", id, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	if err := h.service.Buy(r.Context(), id, req); err != nil {
		switch {
		case errors.Is(err, services.ErrTripNotFound):
			h.log.Warnw("Тур не найден при покупке", "id", id)
			helpers.Error(w, http.StatusNotFound, "Тур не найден")
		default:
			h.log.Errorw("Ошибка при покупке тура", "id", id, "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Не удалось обработать покупку")
		}
		return
	}

	h.log.Infow("Заявка на тур успешно обработана", "id", id, "name", req.UserName, "phone", req.UserPhone)
	helpers.JSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Заявка успешно отправлена",
	})
}
