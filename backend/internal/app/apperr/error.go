package apperr

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeNotFound   Code = "not_found"
	CodeValidation Code = "validation"
)

type Error struct {
	Code    Code
	Message string // можно показать пользователю
	Field   string // для ошибок валидации
}

func (e *Error) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s: %s", e.Code, e.Field, e.Message)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func ErrNotFound(message string) *Error { return &Error{Code: CodeNotFound, Message: message} }

func ErrValidation(field, message string) *Error {
	return &Error{Code: CodeValidation, Message: message, Field: field}
}

// As возвращает ошибку приложения из цепочки, если она там есть.
func As(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)

	return e, ok
}

func Is(err error, code Code) bool {
	e, ok := As(err)
	return ok && e.Code == code
}
