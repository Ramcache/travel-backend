package models

import (
	"database/sql"
	"time"
)

// API модель
type HotelRequest struct {
	Name         string  `json:"name"`
	City         string  `json:"city"`
	Stars        int     `json:"stars"`
	Distance     float64 `json:"distance"`
	DistanceText *string `json:"distance_text"`
	Meals        string  `json:"meals"`
	Guests       *string `json:"guests"`
	PhotoURL     *string `json:"photo_url"`
}

// DB модель
type Hotel struct {
	ID           int
	Name         string
	City         string
	Stars        int
	Distance     float64
	DistanceText sql.NullString
	Meals        string
	Guests       sql.NullString
	PhotoURL     sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// API ответ
type HotelResponse struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	City         string    `json:"city"`
	Stars        int       `json:"stars"`
	Distance     float64   `json:"distance"`
	DistanceText *string   `json:"distance_text"`
	Meals        string    `json:"meals"`
	Guests       *string   `json:"guests"`
	PhotoURL     *string   `json:"photo_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TripHotel struct {
	TripID  int `json:"trip_id"`
	HotelID int `json:"hotel_id"`
	Nights  int `json:"nights"`
}
