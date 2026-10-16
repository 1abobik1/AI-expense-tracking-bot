package expense

import (
	"fmt"

	"expensebot/internal/app/sqlite"
	"expensebot/internal/expenses/models"
)

func entityToModel(e expenseEntity) (models.Expense, error) {
	spentAt, err := models.ParseDate(e.SpentAt[:min(len(e.SpentAt), 10)])
	if err != nil {
		return models.Expense{}, fmt.Errorf("spent_at %q: %w", e.SpentAt, err)
	}

	createdAt, err := sqlite.ParseTime(e.CreatedAt)
	if err != nil {
		return models.Expense{}, fmt.Errorf("created_at %q: %w", e.CreatedAt, err)
	}

	updatedAt, err := sqlite.ParseTime(e.UpdatedAt)
	if err != nil {
		return models.Expense{}, fmt.Errorf("updated_at %q: %w", e.UpdatedAt, err)
	}

	return models.Expense{
		ID:        e.ID,
		UserID:    e.UserID,
		Amount:    models.Money(e.Amount),
		Category:  models.Category(e.Category),
		SpentAt:   spentAt,
		Comment:   e.Comment,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
