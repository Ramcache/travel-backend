package models

import "time"

type TripReview struct {
	ID        int       `json:"id"`
	TripID    int       `json:"trip_id"`
	UserName  string    `json:"user_name"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateReviewRequest struct {
	TripID   int    `json:"trip_id" example:"1"`
	UserName string `json:"user_name" example:"Иван"`
	Rating   int    `json:"rating" example:"5"`
	Comment  string `json:"comment" example:"Отличный тур!"`
}
