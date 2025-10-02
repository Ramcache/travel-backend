package models

import "time"

type TripFilter struct {
	Title         string // поиск по названию тура
	DepartureCity string // город вылета
	TripType      string // тип тура
	Season        string // сезон
	RouteCity     string // город маршрута
	Active        *bool  // статус тура
	StartAfter    time.Time
	EndBefore     time.Time
	Limit         int
	Offset        int
}
