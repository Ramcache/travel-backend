package validators

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

var Validate *validator.Validate

func Init() {
	Validate = validator.New()
}

func TranslateValidationErrors(err error) string {
	if err == nil {
		return ""
	}

	var msgs []string
	for _, e := range err.(validator.ValidationErrors) {
		field := e.Field()
		switch field {
		case "City":
			msgs = append(msgs, "Поле 'Город' обязательно")
		case "Position":
			msgs = append(msgs, "Поле 'Позиция' обязательно")
		default:
			msgs = append(msgs, "Поле '"+field+"' заполнено неверно")
		}
	}
	return strings.Join(msgs, "; ")
}
