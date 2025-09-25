package handlers

import (
	"encoding/json"
	"github.com/Ramcache/travel-backend/internal/services"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type TelegramHandler struct {
	orderService *services.OrderService
	log          *zap.SugaredLogger
}

func NewTelegramHandler(orderService *services.OrderService, log *zap.SugaredLogger) *TelegramHandler {
	return &TelegramHandler{orderService: orderService, log: log}
}

func (h *TelegramHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	var update struct {
		CallbackQuery struct {
			Data string `json:"data"`
		} `json:"callback_query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		h.log.Errorw("Ошибка парсинга Telegram update", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if update.CallbackQuery.Data == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(update.CallbackQuery.Data, "_")
	if len(parts) != 2 {
		w.WriteHeader(http.StatusOK)
		return
	}
	action, idStr := parts[0], parts[1]
	orderID, _ := strconv.Atoi(idStr)

	var status string
	switch action {
	case "confirm":
		status = "confirmed"
	case "reject":
		status = "rejected"
	default:
		status = "pending"
	}

	if err := h.orderService.UpdateStatus(r.Context(), orderID, status); err != nil {
		h.log.Errorw("Ошибка обновления статуса заказа", "id", orderID, "status", status, "err", err)
	}

	w.WriteHeader(http.StatusOK)
}
