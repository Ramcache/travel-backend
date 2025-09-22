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
	Currency        string     `json:"currency"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	BookingDeadline *time.Time `json:"booking_deadline"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateTripRequest struct {
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	PhotoURL        string  `json:"photo_url"`
	DepartureCity   string  `json:"departure_city"`
	TripType        string  `json:"trip_type"`
	Season          string  `json:"season"`
	Price           float64 `json:"price"`
	Currency        string  `json:"currency"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	BookingDeadline string  `json:"booking_deadline"`
}

type UpdateTripRequest struct {
	Title           *string  `json:"title,omitempty"`
	Description     *string  `json:"description,omitempty"`
	PhotoURL        *string  `json:"photo_url,omitempty"`
	DepartureCity   *string  `json:"departure_city,omitempty"`
	TripType        *string  `json:"trip_type,omitempty"`
	Season          *string  `json:"season,omitempty"`
	Price           *float64 `json:"price,omitempty"`
	Currency        *string  `json:"currency,omitempty"`
	StartDate       *string  `json:"start_date,omitempty"`
	EndDate         *string  `json:"end_date,omitempty"`
	BookingDeadline *string  `json:"booking_deadline,omitempty"`
}
