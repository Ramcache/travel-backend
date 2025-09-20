package handlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/services"
)

type CurrencyHandler struct {
	service *services.CurrencyService
	log     *zap.SugaredLogger
}

func NewCurrencyHandler(s *services.CurrencyService, log *zap.SugaredLogger) *CurrencyHandler {
	return &CurrencyHandler{service: s, log: log}
}

// GetRates
// @Summary Get currency rates
// @Tags currency
// @Produce json
// @Success 200 {object} services.CurrencyRate
// @Router /currency [get]
func (h *CurrencyHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	rates, err := h.service.GetRates()
	if err != nil {
		h.log.Errorw("currency_fetch_failed", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "failed to fetch rates")
		return
	}
	helpers.JSON(w, http.StatusOK, rates)
}
