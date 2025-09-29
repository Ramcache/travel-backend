package models

import "time"

type TripRoute struct {
	ID        int       `json:"id"`
	TripID    int       `json:"trip_id"`
	City      string    `json:"city"`
	Transport string    `json:"transport,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TripRouteRequest struct {
	City      string `json:"city" validate:"required"`
	Transport string `json:"transport,omitempty"` // plane, bus, transfer
	Duration  string `json:"duration,omitempty"`  // "6 часов"
	Position  int    `json:"position" validate:"required"`
}
