package models

import "time"

// Табличная модель (из БД)
type TripRoute struct {
	ID        int       `json:"id"`
	TripID    int       `json:"trip_id"`
	City      string    `json:"city"`
	Transport string    `json:"transport,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	StopTime  string    `json:"stop_time,omitempty"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TripRouteRequest struct {
	City      string `json:"city" validate:"required"`
	Transport string `json:"transport,omitempty"`
	Duration  string `json:"duration,omitempty"`
	StopTime  string `json:"stop_time,omitempty"`
	Position  int    `json:"position" validate:"required"`
}

// Старый фронтовый ответ (совместимость)
type TripRouteSegment struct {
	City      string `json:"city"`
	Transport string `json:"transport,omitempty"`
	Duration  string `json:"duration,omitempty"`
	StopTime  string `json:"stop_time,omitempty"`
}

type TripRouteResponse struct {
	Route         []TripRouteSegment `json:"route"`
	TotalDuration string             `json:"total_duration"`
}

// Новый UI-ответ для плашки
// items: city → leg → city → leg → city
type TripRouteUIItem struct {
	Kind         string `json:"kind"`                     // "city" или "leg"
	City         string `json:"city,omitempty"`           // для kind=city
	Transport    string `json:"transport,omitempty"`      // для kind=leg (airplane/bus/train/...)
	DurationText string `json:"duration_text,omitempty"`  // для kind=leg
	StopTimeText string `json:"stop_time_text,omitempty"` // для kind=city (время пересадки)
}

type TripRouteUIResponse struct {
	From                 string            `json:"from"`
	To                   string            `json:"to"`
	Items                []TripRouteUIItem `json:"items"`
	TotalDurationText    string            `json:"total_duration"`
	TotalDurationMinutes int               `json:"total_duration_minutes"`
}

type TripRouteBatchRequest struct {
	Routes []TripRouteRequest `json:"routes" validate:"required,dive"`
}
