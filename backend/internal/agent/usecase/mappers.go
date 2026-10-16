package usecase

import "expensebot/internal/expenses/models"

// Ответ модели → модели трат. Невалидные поля ИИ заменяются безопасными значениями.

func mapItemToExpense(item NewItem, today models.Date) models.Expense {
	return models.Expense{
		Amount:   models.MoneyFromRubles(item.Amount),
		Category: categoryOr(item.Category, models.CategoryOther),
		SpentAt:  dateOr(item.Date, today),
		Comment:  item.Comment,
	}
}

func mapChangesToUpdate(c Changes) models.UpdateExpenseIn {
	var out models.UpdateExpenseIn
	if c.Amount != 0 {
		m := models.MoneyFromRubles(c.Amount)
		out.Amount = &m
	}

	if category, ok := models.ParseCategory(c.Category); ok {
		out.Category = &category
	}

	if d, err := models.ParseDate(c.Date); err == nil {
		out.SpentAt = &d
	}

	if c.Comment != "" {
		out.Comment = &c.Comment
	}

	return out
}

func categoryOr(s string, fallback models.Category) models.Category {
	if c, ok := models.ParseCategory(s); ok {
		return c
	}

	return fallback
}

func dateOr(s string, fallback models.Date) models.Date {
	if d, err := models.ParseDate(s); err == nil {
		return d
	}

	return fallback
}
