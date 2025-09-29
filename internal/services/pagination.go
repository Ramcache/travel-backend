package services

// PaginatedResponse универсальная структура ответа со списками
// Используется в хендлерах, чтобы фронт знал total + items
type PaginatedResponse[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}
