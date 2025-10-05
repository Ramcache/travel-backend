package models

type BuyRequest struct {
	Name      string `json:"name"`
	Date      string `json:"date"`
	Price     string `json:"price"`
	UserName  string `json:"username"`
	UserPhone string `json:"phone"`
}
