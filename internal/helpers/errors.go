package helpers

import "fmt"

// ErrInvalidInput используется для ошибок валидации / некорректных данных.
type InvalidInputError struct {
	msg string
}

func (e *InvalidInputError) Error() string { return e.msg }

// Конструктор
func ErrInvalidInput(msg string) error {
	return &InvalidInputError{msg: msg}
}

// Проверка
func IsInvalidInput(err error) bool {
	_, ok := err.(*InvalidInputError)
	return ok
}

// Пример: ErrInvalidInputf("поле %s обязательно", "email")
func ErrInvalidInputf(format string, args ...any) error {
	return &InvalidInputError{msg: fmt.Sprintf(format, args...)}
}
