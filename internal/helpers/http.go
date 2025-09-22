package helpers

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
)

var ruDict = map[string]string{
	// Общие
	"internal_error":    "Внутренняя ошибка сервера",
	"bad_request":       "Некорректный запрос",
	"unauthorized":      "Требуется авторизация",
	"forbidden":         "Доступ запрещён",
	"not_found":         "Ресурс не найден",
	"conflict":          "Конфликт данных",
	"validation_failed": "Ошибка валидации данных",
	// Доменные (пример)
	"user_not_found":      "Пользователь не найден",
	"invalid_credentials": "Неверный логин или пароль",
	"news_not_found":      "Новость не найдена",
	"duplicate_slug":      "Такой слаг уже используется",
}

func Translate(code string) string {
	return ruDict[code]
}

func FieldsFromValidationErr(err error) map[string]string {
	if err == nil {
		return nil
	}
	out := make(map[string]string)
	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range verrs {
			// Короткие русские сообщения по тегам
			switch fe.Tag() {
			case "required":
				out[fe.Field()] = "Обязательное поле"
			case "email":
				out[fe.Field()] = "Некорректный email"
			case "min":
				out[fe.Field()] = "Слишком короткое значение"
			case "max":
				out[fe.Field()] = "Слишком длинное значение"
			default:
				out[fe.Field()] = "Некорректное значение"
			}
		}
		return out
	}
	// невалидаторная ошибка — как общая
	return map[string]string{"_error": "Некорректные данные"}
}

type mappedErr struct {
	status int
	code   string
	msg    string
}

func MapPgErr(err error) (int, string, string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return http.StatusConflict, "conflict", "Запись с такими данными уже существует"
		case "23503": // foreign_key_violation
			return http.StatusBadRequest, "bad_request", "Нарушена ссылочная целостность"
		}
	}
	return http.StatusInternalServerError, "internal_error", ""
}
