package models

import (
	"database/sql"
	"time"
)

type Order struct {
	ID        int           `json:"id"`
	TripID    sql.NullInt32 `json:"trip_id"`
	UserName  string        `json:"user_name"`
	UserPhone string        `json:"user_phone"`
	Status    string        `json:"status"`
	IsRead    bool          `json:"is_read"`
	CreatedAt time.Time     `json:"created_at"`
}
