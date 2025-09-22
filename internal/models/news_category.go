package models

import "time"

type NewsCategory struct {
	ID        int       `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateNewsCategoryRequest struct {
	Slug  string `json:"slug" example:"hajj_news"`
	Title string `json:"title" example:"Новости хаджа"`
}

type UpdateNewsCategoryRequest struct {
	Slug  *string `json:"slug,omitempty"`
	Title *string `json:"title,omitempty"`
}
