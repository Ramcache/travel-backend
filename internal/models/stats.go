package models

type KV struct {
	Key   string `json:"key" example:"admin"`
	Count int64  `json:"count" example:"42"`
}

type Stats struct {
	TotalUsers     int64 `json:"total_users" example:"100"`
	TotalNews      int64 `json:"total_news" example:"25"`
	TotalTrips     int64 `json:"total_trips" example:"12"`
	UsersByRole    []KV  `json:"users_by_role"`
	NewsByStatus   []KV  `json:"news_by_status"`
	NewsByCategory []KV  `json:"news_by_category"`
	TripsByType    []KV  `json:"trips_by_type"`
	TripsByCity    []KV  `json:"trips_by_city"`
}
