package models

import "time"

// ======== Основная модель тура ========
type Trip struct {
	ID              int                 `json:"id"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	URLs            []string            `json:"urls"` // 👈 массив ссылок
	DepartureCity   string              `json:"departure_city"`
	TripType        string              `json:"trip_type"`
	Season          string              `json:"season"`
	Price           float64             `json:"price"`
	FinalPrice      float64             `json:"final_price"`
	DiscountPercent int                 `json:"discount_percent"`
	Currency        string              `json:"currency"`
	Main            bool                `json:"main"`
	Active          bool                `json:"active"`
	ViewsCount      int                 `json:"views_count"`
	BuysCount       int                 `json:"buys_count"`
	StartDate       time.Time           `json:"start_date"`
	EndDate         time.Time           `json:"end_date"`
	BookingDeadline *time.Time          `json:"booking_deadline"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	Hotels          []TripHotelWithInfo `json:"hotels,omitempty"`
}

// ======== Вспомогательные модели ========
type TripHotelWithInfo struct {
	HotelID  int     `json:"hotel_id"`
	Name     string  `json:"name"`
	City     string  `json:"city"`
	Distance float64 `json:"distance"`
	Meals    string  `json:"meals"`
	Rating   int     `json:"rating"`
	Nights   int     `json:"nights"`
}

// Простая привязка отеля к туру
type HotelAttach struct {
	HotelID int `json:"hotel_id"`
	Nights  int `json:"nights"`
}

// ======== API-запросы ========

// --- Создание тура (используется и в POST, и в PUT) ---
type CreateTripRequest struct {
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	URLs            []string      `json:"urls"` // 👈 массив ссылок
	DepartureCity   string        `json:"departure_city"`
	TripType        string        `json:"trip_type"`
	Season          string        `json:"season"`
	Price           float64       `json:"price"`
	DiscountPercent int           `json:"discount_percent"`
	Currency        string        `json:"currency"`
	Main            bool          `json:"main"`
	Active          bool          `json:"active"`
	StartDate       string        `json:"start_date"`
	EndDate         string        `json:"end_date"`
	BookingDeadline string        `json:"booking_deadline"`
	Hotels          []HotelAttach `json:"hotels,omitempty"`
}

// --- Обновление тура ---
type UpdateTripRequest struct {
	Title           *string       `json:"title,omitempty"`
	Description     *string       `json:"description,omitempty"`
	URLs            *[]string     `json:"urls,omitempty"`
	DepartureCity   *string       `json:"departure_city,omitempty"`
	TripType        *string       `json:"trip_type,omitempty"`
	Season          *string       `json:"season,omitempty"`
	Price           *float64      `json:"price,omitempty"`
	DiscountPercent *int          `json:"discount_percent,omitempty"`
	Currency        *string       `json:"currency,omitempty"`
	Main            *bool         `json:"main,omitempty"`
	Active          *bool         `json:"active,omitempty"`
	StartDate       *string       `json:"start_date,omitempty"`
	EndDate         *string       `json:"end_date,omitempty"`
	BookingDeadline *string       `json:"booking_deadline,omitempty"`
	Hotels          []HotelAttach `json:"hotels,omitempty"`
}

// --- Полное создание тура (тур + отели + маршруты) ---
type CreateTourRequest struct {
	Trip        CreateTripRequest        `json:"trip"`
	Hotels      []HotelRequest           `json:"hotels,omitempty"`
	Routes      []TripRouteRequest       `json:"routes,omitempty"`
	RouteCities map[string]TripRouteCity `json:"route_cities,omitempty"`
}

// --- Полное обновление тура (тур + отели + маршруты) ---
type UpdateTourRequest struct {
	Trip        UpdateTripRequest        `json:"trip"`
	Hotels      []HotelRequest           `json:"hotels,omitempty"`
	Routes      []TripRouteRequest       `json:"routes,omitempty"`
	RouteCities map[string]TripRouteCity `json:"route_cities,omitempty"`
}

// ======== API-ответы ========

type CreateTourResponse struct {
	Success bool            `json:"success"`
	Trip    *Trip           `json:"trip"`
	Hotels  []HotelResponse `json:"hotels"`
	Routes  []TripRoute     `json:"routes"`
}

// Тур с отелями и маршрутами
type TripWithRelations struct {
	Trip   Trip                     `json:"trip"`
	Hotels []HotelResponse          `json:"hotels"`
	Routes *TripRouteCitiesResponse `json:"routes"`
}

// Полный ответ (тур + отели + маршруты)
type TripFullResponse struct {
	Trip   Trip            `json:"trip"`
	Hotels []HotelResponse `json:"hotels"`
	Routes []TripRoute     `json:"routes"`
}

// ======== Методы ========

func (t *Trip) CalculateFinalPrice() {
	if t.DiscountPercent > 0 {
		t.FinalPrice = t.Price * (100 - float64(t.DiscountPercent)) / 100
	} else {
		t.FinalPrice = t.Price
	}
}
