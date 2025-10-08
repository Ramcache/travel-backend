package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ReviewHandler struct {
	service *services.ReviewService
	log     *zap.SugaredLogger
}

func NewReviewHandler(service *services.ReviewService, log *zap.SugaredLogger) *ReviewHandler {
	return &ReviewHandler{service: service, log: log}
}

// Create
// @Summary Leave review
// @Description Добавить отзыв к туру
// @Tags Public — Reviews
// @Accept json
// @Produce json
// @Param request body models.CreateReviewRequest true "Review"
// @Success 200 {object} models.TripReview
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /trips/{trip_id}/reviews [post]
func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}
	tripID, _ := strconv.Atoi(chi.URLParam(r, "trip_id"))
	req.TripID = tripID

	rev, err := h.service.Create(r.Context(), req)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "Не удалось добавить отзыв")
		return
	}

	helpers.JSON(w, http.StatusOK, rev)
}

// ListByTrip
// @Summary List reviews for trip
// @Description Получить список отзывов по туру
// @Tags Public — Reviews
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} models.PaginatedTripReviews
// @Failure 500 {object} helpers.ErrorData
// @Router /trips/{trip_id}/reviews [get]
func (h *ReviewHandler) ListByTrip(w http.ResponseWriter, r *http.Request) {
	tripID, _ := strconv.Atoi(chi.URLParam(r, "trip_id"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	items, total, err := h.service.ListByTrip(r.Context(), tripID, limit, offset)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить отзывы")
		return
	}

	resp := services.PaginatedResponse[models.TripReview]{Total: total, Items: items}
	helpers.JSON(w, http.StatusOK, resp)
}
