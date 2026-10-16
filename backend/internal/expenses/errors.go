package expenses

import "expensebot/internal/app/apperr"

// ErrExpenseNotFound одна на «нет» и «чужое»: чужая трата неотличима от несуществующей.
func ErrExpenseNotFound() *apperr.Error { return apperr.ErrNotFound("трата не найдена") }

func ErrUnknownMember() *apperr.Error {
	return apperr.ErrValidation("user_id", "нет такого члена семьи")
}

func ErrFutureDate() *apperr.Error {
	return apperr.ErrValidation("spent_at", "дата траты в будущем")
}
