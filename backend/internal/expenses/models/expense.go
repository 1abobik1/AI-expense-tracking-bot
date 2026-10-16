package models

import (
	"strings"
	"time"
	"unicode/utf8"

	"expensebot/internal/app/apperr"
)

const MaxCommentLength = 200

type Expense struct {
	ID        int64
	UserID    int64 // Telegram ID владельца
	Amount    Money
	Category  Category
	SpentAt   Date
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate не проверяет дату относительно «сегодня»: это делает юзкейс, у него есть часы.
func (e *Expense) Validate() error {
	e.Comment = strings.TrimSpace(e.Comment)
	switch {
	case e.UserID <= 0:
		return apperr.ErrValidation("user_id", "не указан владелец")
	case e.Amount <= 0:
		return apperr.ErrValidation("amount", "сумма должна быть больше нуля")
	case e.Amount > MaxMoney:
		return apperr.ErrValidation("amount", "слишком большая сумма")
	case !e.Category.Valid():
		return apperr.ErrValidation("category", "неизвестная категория")
	case e.SpentAt.IsZero():
		return apperr.ErrValidation("spent_at", "не указана дата")
	case utf8.RuneCountInString(e.Comment) > MaxCommentLength:
		return apperr.ErrValidation("comment", "комментарий длиннее 200 символов")
	}

	return nil
}

// UpdateExpenseIn: nil-поля не меняются.
type UpdateExpenseIn struct {
	Amount   *Money
	Category *Category
	SpentAt  *Date
	Comment  *string
}

func (p UpdateExpenseIn) IsEmpty() bool {
	return p.Amount == nil && p.Category == nil && p.SpentAt == nil && p.Comment == nil
}

func (e *Expense) Apply(p UpdateExpenseIn) {
	if p.Amount != nil {
		e.Amount = *p.Amount
	}

	if p.Category != nil {
		e.Category = *p.Category
	}

	if p.SpentAt != nil {
		e.SpentAt = *p.SpentAt
	}

	if p.Comment != nil {
		e.Comment = *p.Comment
	}
}

// GetExpensesParams: нулевые поля не фильтруют.
type GetExpensesParams struct {
	UserID   int64 // учитывается только в AdminScope
	From, To Date  // включительно
	Category Category
	Limit    int
	Offset   int
}

// Summary считается без учёта Limit/Offset.
type Summary struct {
	Count      int
	Total      Money
	ByCategory map[Category]Money
}
