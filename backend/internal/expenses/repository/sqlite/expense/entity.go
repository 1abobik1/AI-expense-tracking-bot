package expense

// expenseEntity — строка таблицы expenses. Даты и время в SQLite/D1 — TEXT.
type expenseEntity struct {
	ID        int64
	UserID    int64
	Amount    int64 // копейки
	Category  string
	SpentAt   string
	Comment   string
	CreatedAt string
	UpdatedAt string
}

func (e *expenseEntity) fields() []any {
	return []any{&e.ID, &e.UserID, &e.Amount, &e.Category, &e.SpentAt, &e.Comment, &e.CreatedAt, &e.UpdatedAt}
}
