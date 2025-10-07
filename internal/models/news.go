package models

import "time"

// ======== Основная модель новости ========
type News struct {
	ID            int       `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Excerpt       string    `json:"excerpt"`
	Content       string    `json:"content"`
	CategoryID    *int      `json:"category_id,omitempty"`
	MediaType     string    `json:"media_type"`
	URLs          []string  `json:"urls"` // 👈 массив ссылок вместо preview_url
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

// ======== Параметры фильтрации ========
type ListNewsParams struct {
	CategoryID int    `json:"category_id"`
	MediaType  string `json:"media_type"`
	Search     string `json:"search"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Status     string `json:"status"`
}

// ======== API-запросы ========

// Создание новости
type CreateNewsRequest struct {
	Title       string   `json:"title"`
	Excerpt     string   `json:"excerpt"`
	Content     string   `json:"content"`
	CategoryID  int      `json:"category_id"`
	MediaType   string   `json:"media_type"`
	URLs        []string `json:"urls"` // 👈 массив ссылок
	VideoURL    *string  `json:"video_url,omitempty"`
	Status      string   `json:"status"`
	PublishedAt string   `json:"published_at"`
}

// Обновление новости
type UpdateNewsRequest struct {
	Slug        *string   `json:"slug,omitempty"`
	Title       *string   `json:"title,omitempty"`
	Excerpt     *string   `json:"excerpt,omitempty"`
	Content     *string   `json:"content,omitempty"`
	CategoryID  *int      `json:"category_id,omitempty"`
	MediaType   *string   `json:"media_type,omitempty"`
	URLs        *[]string `json:"urls,omitempty"` // 👈 массив ссылок
	VideoURL    *string   `json:"video_url,omitempty"`
	Status      *string   `json:"status,omitempty"`
	PublishedAt *string   `json:"published_at,omitempty"`
}
