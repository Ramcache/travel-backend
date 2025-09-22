package helpers

import (
	"encoding/json"
	"net/http"
)

// Envelope — общий формат ответа
type Envelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}

// ErrorData — структура ошибки (уходит в Data)
type ErrorData struct {
	Code    string            `json:"code"`             // машинный код ошибки
	Message string            `json:"message"`          // человеко-читаемый текст (на русском)
	Fields  map[string]string `json:"fields,omitempty"` // ошибки по полям (для валидации)
}

// JSON — успешный ответ
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: true,
		Data:    payload,
	})
}

// Error — ошибка, всегда в Data
func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: false,
		Data: ErrorData{
			Code:    http.StatusText(status), // можно заменить на кастомный код
			Message: message,
		},
	})
}

// ValidationError — для удобства при StructValidation
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: false,
		Data: ErrorData{
			Code:    "validation_failed",
			Message: "Ошибка валидации данных",
			Fields:  fields,
		},
	})
}
