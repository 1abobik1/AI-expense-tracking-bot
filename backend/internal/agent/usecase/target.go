package usecase

import (
	"context"
	"slices"
	"strings"
	"unicode/utf8"

	"expensebot/internal/expenses/models"
)

const (
	targetSearchDepth = 100
	maxAmbiguous      = 5
)

// findTarget возвращает nil и уточняющий ответ, если трата не нашлась или подходят несколько.
func (a *AgentUseCase) findTarget(ctx context.Context, scope models.Scope, t Target, today models.Date) (*models.Expense, string, error) {
	if !t.Last && t.Text == "" && t.Amount == 0 && t.Date == "" {
		return nil, "Не понял, о какой трате речь. Уточни: «последнюю», «такси», «кофе 350».", nil
	}

	params := models.GetExpensesParams{Limit: targetSearchDepth}
	if d, err := models.ParseDate(t.Date); err == nil {
		params.From, params.To = d, d
	}

	candidates, err := a.expensesUC.GetExpenses(ctx, scope, params)
	if err != nil {
		return nil, "", err
	}

	amount := models.MoneyFromRubles(t.Amount)

	matches := slices.DeleteFunc(candidates, func(e models.Expense) bool {
		return (amount != 0 && e.Amount != amount) || (t.Text != "" && !matchesText(e, t.Text))
	})
	if t.Last && len(matches) > 0 {
		// «Последнее» — последнее добавленное, а не самое позднее по дате.
		last := slices.MaxFunc(matches, func(x, y models.Expense) int { return int(x.ID - y.ID) })
		return &last, "", nil
	}

	switch len(matches) {
	case 0:
		return nil, "Не нашёл такую трату среди твоих последних.", nil
	case 1:
		return &matches[0], "", nil
	}
	var b strings.Builder
	b.WriteString("Подходят несколько трат — уточни сумму или дату:\n")

	for i := range min(len(matches), maxAmbiguous) {
		b.WriteString("• " + formatExpense(matches[i], today) + "\n")
	}

	return nil, strings.TrimSpace(b.String()), nil
}

// matchesText сравнивает по началу слов, чтобы пережить опечатки и окончания.
func matchesText(e models.Expense, text string) bool {
	if strings.Contains(normalize(e.Comment), normalize(text)) || normalize(e.Category.Label()) == normalize(text) {
		return true
	}

	commentStems := stems(e.Comment)
	for _, s := range stems(text) {
		if slices.Contains(commentStems, s) {
			return true
		}
	}

	return false
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), "ё", "е")
}

func stems(s string) []string {
	var out []string
	for _, w := range strings.FieldsFunc(normalize(s), func(r rune) bool { return !isLetter(r) }) {
		if utf8.RuneCountInString(w) < 3 {
			continue
		}

		out = append(out, string([]rune(w)[:min(4, utf8.RuneCountInString(w))]))
	}

	return out
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'а' && r <= 'я') || r == 'ё'
}
