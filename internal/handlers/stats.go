package handlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/services"
)

type StatsHandler struct {
	svc *services.StatsService
	log *zap.SugaredLogger
}

func NewStatsHandler(svc *services.StatsService, log *zap.SugaredLogger) *StatsHandler {
	return &StatsHandler{svc: svc, log: log}
}

// Get
// @Summary Admin statistics
// @Tags admin
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/stats [get]
func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.Get(r.Context())
	if err != nil {
		h.log.Errorw("stats failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось получить статистику")
		return
	}
	helpers.JSON(w, http.StatusOK, data)
}
