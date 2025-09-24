package models

type BuyRequest struct {
	UserName  string `json:"name" example:"Иван"`
	UserPhone string `json:"phone" example:"+79998887766"`
	TripID    int    `json:"trip_id" example:"1"`
}
