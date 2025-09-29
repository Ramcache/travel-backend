package handlers

import (
	"net/http"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/services"
	"go.uber.org/zap"
)

type SearchHandler struct {
	service *services.SearchService
	log     *zap.SugaredLogger
}

func NewSearchHandler(service *services.SearchService, log *zap.SugaredLogger) *SearchHandler {
	return &SearchHandler{service: service, log: log}
}

// GlobalSearch
// @Summary Global search (trips + news)
// @Tags search
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} models.SearchResult
// @Failure 400 {object} helpers.ErrorData "Некорректный запрос"
// @Failure 500 {object} helpers.ErrorData "Ошибка поиска"
// @Router /search [get]
func (h *SearchHandler) GlobalSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.log.Warnw("Поиск без параметра q")
		helpers.Error(w, http.StatusBadRequest, "Не указан поисковый запрос")
		return
	}

	results, err := h.service.GlobalSearch(r.Context(), query)
	if err != nil {
		h.log.Errorw("Ошибка глобального поиска", "query", query, "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось выполнить поиск")
		return
	}

	h.log.Infow("Результаты поиска получены", "query", query, "count", len(results))
	helpers.JSON(w, http.StatusOK, results)
}
