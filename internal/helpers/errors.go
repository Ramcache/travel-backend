package helpers

type InvalidInputError struct {
	msg string
}

func (e *InvalidInputError) Error() string {
	return e.msg
}

func ErrInvalidInput(msg string) error {
	return &InvalidInputError{msg: msg}
}

func IsInvalidInput(err error) bool {
	_, ok := err.(*InvalidInputError)
	return ok
}
