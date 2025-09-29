package models

// PaginatedTripReviews нужен только для Swagger
type PaginatedTripReviews struct {
	Total int          `json:"total"`
	Items []TripReview `json:"items"`
}

// PaginatedOrders нужен только для Swagger
type PaginatedOrders struct {
	Total int     `json:"total"`
	Items []Order `json:"items"`
}

// PaginatedFeedbacks нужен только для Swagger
type PaginatedFeedbacks struct {
	Total int        `json:"total"`
	Items []Feedback `json:"items"`
}
