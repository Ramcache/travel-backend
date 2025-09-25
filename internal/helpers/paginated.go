package helpers

// PaginatedResponse универсальная структура для списков с total
type PaginatedResponse[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}
