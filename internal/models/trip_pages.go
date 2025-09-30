package models

import "time"

// Countdown — удобный вид для фронта
type Countdown struct {
	Days    int `json:"days"`
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

// TripPageResponse — агрегированный ответ для страницы тура
type TripPageResponse struct {
	Trip          Trip                 `json:"trip"`
	Countdown     *Countdown           `json:"countdown,omitempty"`
	DurationDays  int                  `json:"duration_days"`
	Routes        []TripRoute          `json:"routes"`
	Hotels        []HotelResponse      `json:"hotels"`
	Reviews       TripPageReviews      `json:"reviews"`
	PopularTrips  []Trip               `json:"popular_trips"`
	News          []News               `json:"news"`
	CurrencyRates CurrencyRatesPayload `json:"currency_rates"`
}

// TripPageReviews — компактный пагинированный блок (без обобщений для Swagger)
type TripPageReviews struct {
	Total int          `json:"total"`
	Items []TripReview `json:"items"`
}

// CurrencyRatesPayload — пригодный для фронта вид курсов
type CurrencyRatesPayload struct {
	USD float64 `json:"usd"`
	SAR float64 `json:"sar"`
}

// Helper для длительности в днях
func CalcDurationDays(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours())/24 + 1
}
