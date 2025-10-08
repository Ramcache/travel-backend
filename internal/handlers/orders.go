package handlers

import (
	"errors"
	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/repository"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type OrderHandler struct {
	service *services.OrderService
	log     *zap.SugaredLogger
}

func NewOrderHandler(service *services.OrderService, log *zap.SugaredLogger) *OrderHandler {
	return &OrderHandler{service: service, log: log}
}

// List
// @Summary Get orders list (admin)
// @Description Получить список заказов с пагинацией и фильтрацией
// @Tags Admin — Orders
// @Security Bearer
// @Produce json
// @Param limit query int false "Количество записей (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Param status query string false "Фильтр по статусу (new/in_progress/done/canceled)"
// @Param phone query string false "Фильтр по телефону"
// @Param is_read query bool false "Фильтр по прочитанности"
// @Success 200 {object} services.OrdersWithTotal
// @Failure 500 {object} helpers.ErrorData "Не удалось получить список заказов"
// @Router /admin/orders [get]
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	status := q.Get("status")
	phone := q.Get("phone")

	var isRead *bool
	if v := q.Get("is_read"); v != "" {
		b := v == "true" || v == "1"
		isRead = &b
	}

	result, err := h.service.List(r.Context(), limit, offset, status, phone, isRead)
	if err != nil {
		h.log.Errorw("Ошибка получения списка заказов", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить список заказов")
		return
	}

	h.log.Infow("Список заказов успешно получен", "count", len(result.Orders), "total", result.Total)
	helpers.JSON(w, http.StatusOK, result)
}

// UpdateStatus
// @Summary Update order status
// @Description Обновить статус заказа (admin)
// @Tags Admin — Orders
// @Security Bearer
// @Param id path int true "Order ID"
// @Param status query string true "Новый статус (new/in_progress/done/canceled)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData "Некорректные данные"
// @Failure 404 {object} helpers.ErrorData "Заказ не найден"
// @Failure 500 {object} helpers.ErrorData "Не удалось обновить статус"
// @Router /admin/orders/{id}/status [post]
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warnw("Некорректный ID заказа", "id", idStr, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		helpers.Error(w, http.StatusBadRequest, "Не указан статус")
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, status); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.log.Warnw("Заказ не найден при обновлении статуса", "id", id)
			helpers.Error(w, http.StatusNotFound, "Заказ не найден")
			return
		}
		h.log.Errorw("Ошибка при обновлении статуса заказа", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить статус")
		return
	}

	h.log.Infow("Статус заказа обновлён", "id", id, "status", status)
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// MarkAsRead
// @Summary Mark order as read
// @Description Пометить заказ как прочитанный
// @Tags Admin — Orders
// @Security Bearer
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData "Некорректный ID"
// @Failure 500 {object} helpers.ErrorData "Не удалось обновить заказ"
// @Router /admin/orders/{id}/read [post]
func (h *OrderHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warnw("Некорректный ID заказа", "id", idStr, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	if err := h.service.MarkAsRead(r.Context(), id); err != nil {
		h.log.Errorw("Ошибка при отметке заказа как прочитанного", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось обновить заказ")
		return
	}

	h.log.Infow("Заказ помечен как прочитанный", "id", id)
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Delete order
// @Summary Delete order
// @Tags Admin — Orders
// @Security Bearer
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorData
// @Failure 404 {object} helpers.ErrorData
// @Failure 500 {object} helpers.ErrorData
// @Router /admin/orders/{id} [delete]
func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warnw("Некорректный ID заказа", "id", idStr, "err", err)
		helpers.Error(w, http.StatusBadRequest, "Некорректный ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.log.Warnw("Заказ не найден при удалении", "id", id)
			helpers.Error(w, http.StatusNotFound, "Заказ не найден")
			return
		}
		h.log.Errorw("Ошибка при удалении заказа", "id", id, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось удалить заказ")
		return
	}

	h.log.Infow("Заказ удалён", "id", id)
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
