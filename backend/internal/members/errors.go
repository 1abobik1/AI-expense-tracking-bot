package members

import "expensebot/internal/app/apperr"

func ErrMemberNotFound() *apperr.Error {
	return apperr.ErrNotFound("участник не найден")
}

func ErrUnknownStatus() *apperr.Error {
	return apperr.ErrValidation("status", "неизвестный статус")
}

func ErrNameTooLong() *apperr.Error {
	return apperr.ErrValidation("name", "имя длиннее 50 символов")
}
