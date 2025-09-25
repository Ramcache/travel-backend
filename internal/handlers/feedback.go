package handlers

import (
	"encoding/json"
	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type FeedbackHandler struct {
	service *services.FeedbackService
	log     *zap.SugaredLogger
}

func NewFeedbackHandler(service *services.FeedbackService, log *zap.SugaredLogger) *FeedbackHandler {
	return &FeedbackHandler{service: service, log: log}
}

// Create
// Public: Feedback form
// @Summary Feedback form
// @Description Оставить заявку "Перезвоните мне"
// @Tags public
// @Accept json
// @Produce json
// @Param request body models.FeedbackRequest true "Имя и телефон"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /feedback [post]
func (h *FeedbackHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректное тело запроса")
		return
	}

	if err := h.service.Create(r.Context(), req); err != nil {
		h.log.Errorw("Ошибка при сохранении feedback", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось отправить заявку")
		return
	}

	h.log.Infow("Заявка на консультацию отправлена", "name", req.UserName, "phone", req.UserPhone)
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// List feedbacks
// @Summary Get feedbacks
// @Description Получить список заявок (админка)
// @Tags admin
// @Security Bearer
// @Produce json
// @Param limit query int false "Количество (20)"
// @Param offset query int false "Смещение (0)"
// @Success 200 {array} models.Feedback
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/feedbacks [get]
func (h *FeedbackHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	list, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить список заявок")
		return
	}

	helpers.JSON(w, http.StatusOK, list)
}

// MarkAsRead
// @Summary Mark feedback as read
// @Tags admin
// @Security Bearer
// @Param id path int true "Feedback ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/feedbacks/{id}/read [post]
func (h *FeedbackHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	if err := h.service.MarkAsRead(r.Context(), id); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить заявку")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
