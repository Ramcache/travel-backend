package models

type SearchResult struct {
	Type        string `json:"type"`
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Date        string `json:"date"`
	TripType    string `json:"trip_type,omitempty"`
	Highlighted bool   `json:"highlighted"`
}
