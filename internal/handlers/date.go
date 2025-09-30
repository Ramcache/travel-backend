package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/hablullah/go-hijri"
	"go.uber.org/zap"
)

type DateHandler struct {
	log *zap.SugaredLogger
}

func NewDateHandler(log *zap.SugaredLogger) *DateHandler {
	return &DateHandler{log: log}
}

// Месяцы по-русски (григорианский)
var gregorianMonths = map[time.Month]string{
	time.January:   "января",
	time.February:  "февраля",
	time.March:     "марта",
	time.April:     "апреля",
	time.May:       "мая",
	time.June:      "июня",
	time.July:      "июля",
	time.August:    "августа",
	time.September: "сентября",
	time.October:   "октября",
	time.November:  "ноября",
	time.December:  "декабря",
}

// Месяцы по-русски (хиджра, Umm al-Qura)
var hijriMonths = map[int]string{
	1:  "мухаррам",
	2:  "сафар",
	3:  "раб-и-авваль",
	4:  "раб-и-ахир",
	5:  "джумада-уль-уля",
	6:  "джумада-уль-ахира",
	7:  "раджаб",
	8:  "шаабан",
	9:  "рамадан",
	10: "зуль-када",
	11: "зуль-хиджа",
	12: "зуль-хиджа",
}

// Today
// Public: Get today date
// @Summary Get today's date
// @Description Получить сегодняшнюю дату в григорианском и исламском календарях
// @Tags public
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} helpers.ErrorData
// @Router /date/today [get]
func (h *DateHandler) Today(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	gregorian := fmt.Sprintf("%d %s %d", now.Day(), gregorianMonths[now.Month()], now.Year())

	hDate, err := hijri.CreateUmmAlQuraDate(now)
	if err != nil {
		h.log.Errorw("Ошибка конвертации даты", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось сконвертировать дату")
		return
	}

	hijriStr := fmt.Sprintf("%d %s %d", hDate.Day, hijriMonths[int(hDate.Month)], hDate.Year)

	resp := map[string]string{
		"date": fmt.Sprintf("%s / %s", gregorian, hijriStr),
	}

	helpers.JSON(w, http.StatusOK, resp)
}
