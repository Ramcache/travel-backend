package handlers

import (
	"errors"
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
// @Summary Получить курсы валют
// @Tags currency
// @Produce json
// @Success 200 {object} services.CurrencyRate
// @Failure 500 {object} helpers.ErrorData "Ошибка получения курсов валют"
// @Router /currency [get]
func (h *CurrencyHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	rates, err := h.service.GetRates(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrFetchFailed):
			h.log.Errorw("Ошибка при запросе курсов валют", "err", err)
			helpers.Error(w, http.StatusBadGateway, "Сервис курсов валют временно недоступен")
		case errors.Is(err, services.ErrDecodeFailed):
			h.log.Errorw("Ошибка обработки ответа ЦБ РФ", "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Ошибка обработки ответа сервиса валют")
		default:
			h.log.Errorw("Неизвестная ошибка при получении курсов валют", "err", err)
			helpers.Error(w, http.StatusInternalServerError, "Не удалось получить курсы валют")
		}
		return
	}

	h.log.Infow("Курсы валют успешно получены", "USD", rates.USD, "SAR", rates.SAR)
	helpers.JSON(w, http.StatusOK, rates)
}
