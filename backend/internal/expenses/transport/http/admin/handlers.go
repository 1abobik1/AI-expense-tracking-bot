package admin

import (
	"log/slog"
	"net/http"

	"expensebot/internal/expenses"

	apphttp "expensebot/internal/app/http"
)

// Handlers — траты для панели главы семьи: все запросы в AdminScope.
type Handlers struct {
	expensesUC expenses.ExpensesUseCase
	log        *slog.Logger
}

func NewHandlers(expensesUC expenses.ExpensesUseCase, log *slog.Logger) *Handlers {
	return &Handlers{expensesUC: expensesUC, log: log}
}

func (h *Handlers) Register(mux *apphttp.Router, admin func(http.Handler) http.Handler) {
	route := func(pattern string, fn http.HandlerFunc) { mux.Handle(pattern, admin(fn)) }
	route("GET /api/categories", h.getCategories)
	route("GET /api/expenses", h.getExpenses)
	route("POST /api/expenses", h.createExpense)
	route("GET /api/expenses/{id}", h.getExpense)
	route("PATCH /api/expenses/{id}", h.updateExpense)
	route("DELETE /api/expenses/{id}", h.deleteExpense)
}
