package helpers

type InvalidInputError struct {
	msg string
}

func (e *InvalidInputError) Error() string {
	return e.msg
}

// конструктор
func ErrInvalidInput(msg string) error {
	return &InvalidInputError{msg: msg}
}

// проверка (для хендлеров)
func IsInvalidInput(err error) bool {
	_, ok := err.(*InvalidInputError)
	return ok
}
