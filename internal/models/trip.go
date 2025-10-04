package models

import "time"

type Trip struct {
	ID              int        `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	PhotoURL        string     `json:"photo_url"`
	DepartureCity   string     `json:"departure_city"`
	TripType        string     `json:"trip_type"`
	Season          string     `json:"season"`
	Price           float64    `json:"price"`
	FinalPrice      float64    `json:"final_price"`
	DiscountPercent int        `json:"discount_percent"`
	Currency        string     `json:"currency"`
	Main            bool       `json:"main"`
	Active          bool       `json:"active"`
	ViewsCount      int        `json:"views_count"`
	BuysCount       int        `json:"buys_count"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	BookingDeadline *time.Time `json:"booking_deadline"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	Hotels []TripHotelWithInfo `json:"hotels,omitempty"`
}

type TripHotelWithInfo struct {
	HotelID  int     `json:"hotel_id"`
	Name     string  `json:"name"`
	City     string  `json:"city"`
	Distance float64 `json:"distance"`
	Meals    string  `json:"meals"`
	Rating   int     `json:"rating"`
	Nights   int     `json:"nights"`
}

type HotelAttach struct {
	HotelID int `json:"hotel_id"`
	Nights  int `json:"nights"`
}

type CreateTripRequest struct {
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	PhotoURL        string        `json:"photo_url"`
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

type UpdateTripRequest struct {
	Title           *string       `json:"title,omitempty"`
	Description     *string       `json:"description,omitempty"`
	PhotoURL        *string       `json:"photo_url,omitempty"`
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

type CreateTourRequest struct {
	Trip        CreateTripRequest        `json:"trip"`
	Hotels      []HotelRequest           `json:"hotels"`
	RouteCities map[string]TripRouteCity `json:"route_cities"`
}

type CreateTourResponse struct {
	Success bool            `json:"success"`
	Trip    *Trip           `json:"trip"`
	Hotels  []HotelResponse `json:"hotels"`
	Routes  []TripRoute     `json:"routes"`
}

// TripWithRelations — тур с отелями и маршрутом
type TripWithRelations struct {
	Trip   Trip                     `json:"trip"`
	Hotels []HotelResponse          `json:"hotels"`
	Routes *TripRouteCitiesResponse `json:"routes"`
}

func (t *Trip) CalculateFinalPrice() {
	if t.DiscountPercent > 0 {
		t.FinalPrice = t.Price * (100 - float64(t.DiscountPercent)) / 100
	} else {
		t.FinalPrice = t.Price
	}
}

type TripFullUpdateRequest struct {
	Trip   UpdateTripRequest `json:"trip"`
	Hotels []TripHotel       `json:"hotels"`
	Routes []TripRoute       `json:"routes"`
}
