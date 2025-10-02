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

var hijriMonths = map[int]string{
	1:  "мухаррам",          // محرّم (запретный)
	2:  "сафар",             // صفر
	3:  "рабиʿаль-авваль",   // ربيع الأول
	4:  "рабиʿас-сани",      // ربيع الثاني
	5:  "джумада аль-уля",   // جمادى الأولى
	6:  "джумада аль-ахира", // جمادى الآخرة
	7:  "раджаб",            // رجب (запретный)
	8:  "шаабан",            // شعبان
	9:  "рамадан",           // رمضان
	10: "шавваль",           // شوّال
	11: "зуль-када",         // ذو القعدة (запретный)
	12: "зуль-хиджа",        // ذو الحجة (запретный)
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

	gregorian := fmt.Sprintf("%d %s", now.Day(), gregorianMonths[now.Month()])

	hDate, err := hijri.CreateUmmAlQuraDate(now)
	if err != nil {
		h.log.Errorw("Ошибка конвертации даты", "err", err)
		helpers.Error(w, http.StatusInternalServerError, "Не удалось сконвертировать дату")
		return
	}

	hijriStr := fmt.Sprintf("%d %s", hDate.Day, hijriMonths[int(hDate.Month)])

	resp := map[string]string{
		"date": fmt.Sprintf("%s / %s", gregorian, hijriStr),
	}

	helpers.JSON(w, http.StatusOK, resp)
}
