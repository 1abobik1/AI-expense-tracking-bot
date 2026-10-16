package expenses

import (
	"context"

	"expensebot/internal/expenses/models"
)

// ExpensesRepo обязан ограничивать каждый запрос переданным Scope.
type ExpensesRepo interface {
	CreateExpense(ctx context.Context, e *models.Expense) error
	GetExpense(ctx context.Context, scope models.Scope, id int64) (*models.Expense, error)
	GetExpenses(ctx context.Context, scope models.Scope, params models.GetExpensesParams) ([]models.Expense, error)
	GetSummary(ctx context.Context, scope models.Scope, params models.GetExpensesParams) (*models.Summary, error)
	UpdateExpense(ctx context.Context, scope models.Scope, e *models.Expense) error
	DeleteExpense(ctx context.Context, scope models.Scope, id int64) error
}
