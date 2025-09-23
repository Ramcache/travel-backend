package helpers

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}

type ErrorData struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: true,
		Data:    payload,
	})
}

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: false,
		Data: ErrorData{
			Code:    http.StatusText(status),
			Message: message,
		},
	})
}

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
