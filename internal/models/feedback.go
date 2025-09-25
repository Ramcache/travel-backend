package models

import "time"

type Feedback struct {
	ID        int       `json:"id"`
	UserName  string    `json:"user_name"`
	UserPhone string    `json:"user_phone"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
