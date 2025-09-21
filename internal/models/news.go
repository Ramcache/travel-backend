package models

import "time"

type News struct {
	ID            int       `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Excerpt       string    `json:"excerpt"`
	Content       string    `json:"content"`
	Category      string    `json:"category"`
	MediaType     string    `json:"media_type"`
	PreviewURL    string    `json:"preview_url"`
	VideoURL      *string   `json:"video_url,omitempty"`
	CommentsCount int       `json:"comments_count"`
	RepostsCount  int       `json:"reposts_count"`
	ViewsCount    int       `json:"views_count"`
	AuthorID      *int      `json:"author_id,omitempty"`
	Status        string    `json:"status"`
	PublishedAt   time.Time `json:"published_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ListNewsParams struct {
	Category  string `json:"category"`
	MediaType string `json:"media_type"`
	Search    string `json:"search"`
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	Status    string `json:"status"`
}

type CreateNewsRequest struct {
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Excerpt     string  `json:"excerpt"`
	Content     string  `json:"content"`
	Category    string  `json:"category"`
	MediaType   string  `json:"media_type"`
	PreviewURL  string  `json:"preview_url"`
	VideoURL    *string `json:"video_url"`
	Status      string  `json:"status"`
	PublishedAt string  `json:"published_at"`
}

type UpdateNewsRequest struct {
	Slug        *string `json:"slug"`
	Title       *string `json:"title"`
	Excerpt     *string `json:"excerpt"`
	Content     *string `json:"content"`
	Category    *string `json:"category"`
	MediaType   *string `json:"media_type"`
	PreviewURL  *string `json:"preview_url"`
	VideoURL    *string `json:"video_url"`
	Status      *string `json:"status"`
	PublishedAt *string `json:"published_at"`
}
