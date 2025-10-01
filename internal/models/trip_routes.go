package models

import (
	"fmt"
	"time"
)

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
	Position  int    `json:"position,omitempty"`
}

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

type TripRouteUIItem struct {
	Kind         string `json:"kind"`
	City         string `json:"city,omitempty"`
	Transport    string `json:"transport,omitempty"`
	DurationText string `json:"duration_text,omitempty"`
	StopTimeText string `json:"stop_time_text,omitempty"`
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

type TripRouteCitiesRequest struct {
	Cities map[string]TripRouteCity `json:"route_cities"`
}

type TripRouteCity struct {
	City      string `json:"city" validate:"required"`
	Transport string `json:"transport,omitempty"`
	Duration  string `json:"duration,omitempty"`
	StopTime  string `json:"stop_time,omitempty"`
}

type TripRouteCitiesResponse struct {
	Cities map[string]TripRouteCity `json:"route_cities"`
}

func ConvertCitiesToRoutes(cities map[string]TripRouteCity) []TripRouteRequest {
	routes := make([]TripRouteRequest, 0, len(cities))
	for i := 1; ; i++ {
		key := fmt.Sprintf("city_%d", i)
		c, ok := cities[key]
		if !ok {
			break
		}
		routes = append(routes, TripRouteRequest{
			City:      c.City,
			Transport: c.Transport,
			Duration:  c.Duration,
			StopTime:  c.StopTime,
			Position:  i,
		})
	}
	return routes
}

func ConvertRoutesToCities(routes []TripRoute) TripRouteCitiesResponse {
	resp := TripRouteCitiesResponse{Cities: make(map[string]TripRouteCity)}
	for i, rt := range routes {
		key := fmt.Sprintf("city_%d", i+1)
		resp.Cities[key] = TripRouteCity{
			City:      rt.City,
			Transport: rt.Transport,
			Duration:  rt.Duration,
			StopTime:  rt.StopTime,
		}
	}
	return resp
}
