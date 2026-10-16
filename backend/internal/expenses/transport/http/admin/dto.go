package admin

import "time"

type expenseOut struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	SpentAt   string    `json:"spent_at"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type createExpenseIn struct {
	UserID   int64   `json:"user_id"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	SpentAt  string  `json:"spent_at"`
	Comment  string  `json:"comment"`
}

// updateExpenseIn — PATCH: отсутствующие поля не меняются.
type updateExpenseIn struct {
	Amount   *float64 `json:"amount"`
	Category *string  `json:"category"`
	SpentAt  *string  `json:"spent_at"`
	Comment  *string  `json:"comment"`
}

type expensesOut struct {
	Items []expenseOut `json:"items"`
	Total totalOut     `json:"total"`
}

type totalOut struct {
	Count      int                `json:"count"`
	Amount     float64            `json:"amount"`
	ByCategory map[string]float64 `json:"by_category"`
}

type categoryOut struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}
