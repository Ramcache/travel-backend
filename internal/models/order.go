package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type NullInt32 struct {
	sql.NullInt32
}

func (n NullInt32) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Int32)
}

func (n *NullInt32) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		n.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &n.Int32)
	n.Valid = (err == nil)
	return err
}

type Order struct {
	ID int `json:"id"`
	// TripID может быть числом или null
	// swagger:strfmt int32
	TripID    NullInt32 `json:"trip_id" swaggertype:"integer" example:"123"`
	UserName  string    `json:"user_name"`
	UserPhone string    `json:"user_phone"`
	Status    string    `json:"status"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
