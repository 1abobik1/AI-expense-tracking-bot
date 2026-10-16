package expenses

import (
	"context"

	"expensebot/internal/expenses/models"
)

// ExpensesUseCase общий для бота и панели, разница только в Scope.
type ExpensesUseCase interface {
	Today() models.Date
	CreateExpense(ctx context.Context, scope models.Scope, e models.Expense) (*models.Expense, error)
	GetExpense(ctx context.Context, scope models.Scope, id int64) (*models.Expense, error)
	GetExpenses(ctx context.Context, scope models.Scope, params models.GetExpensesParams) ([]models.Expense, error)
	GetSummary(ctx context.Context, scope models.Scope, params models.GetExpensesParams) (*models.Summary, error)
	UpdateExpense(ctx context.Context, scope models.Scope, id int64, in models.UpdateExpenseIn) (*models.Expense, error)
	DeleteExpense(ctx context.Context, scope models.Scope, id int64) (*models.Expense, error)
}
