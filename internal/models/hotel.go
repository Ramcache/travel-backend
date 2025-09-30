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

func ToHotelResponses(hotels []Hotel) []HotelResponse {
	resp := make([]HotelResponse, 0, len(hotels))
	for _, h := range hotels {
		var distanceText, guests, photoURL *string
		if h.DistanceText.Valid {
			distanceText = &h.DistanceText.String
		}
		if h.Guests.Valid {
			guests = &h.Guests.String
		}
		if h.PhotoURL.Valid {
			photoURL = &h.PhotoURL.String
		}

		resp = append(resp, HotelResponse{
			ID:           h.ID,
			Name:         h.Name,
			City:         h.City,
			Stars:        h.Stars,
			Distance:     h.Distance,
			DistanceText: distanceText,
			Meals:        h.Meals,
			Guests:       guests,
			PhotoURL:     photoURL,
			CreatedAt:    h.CreatedAt,
			UpdatedAt:    h.UpdatedAt,
		})
	}
	return resp
}
