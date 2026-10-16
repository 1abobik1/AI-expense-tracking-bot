package usecase

import (
	"slices"
	"strconv"
	"strings"

	"expensebot/internal/expenses/models"
)

const listPreview = 10

const HelpText = `Я записываю траты. Пиши обычным текстом:
• «кофе 350», «такси 1200 вчера» — добавить
• «сколько потратил на еду в этом месяце», «покажи траты за неделю» — посмотреть
• «последнее было не 350, а 450» — исправить
• «удали такси» — удалить`

func itoa(n int) string { return strconv.Itoa(n) }

func formatExpense(e models.Expense, today models.Date) string {
	name := e.Comment
	if name == "" {
		name = e.Category.Label()
	}

	return name + " — " + e.Amount.String() + ", " + e.Category.Label() + ", " + humanDate(e.SpentAt, today)
}

func formatAdded(created []models.Expense, today models.Date) string {
	if len(created) == 1 {
		return "✅ Записал: " + formatExpense(created[0], today)
	}
	var b strings.Builder
	b.WriteString("✅ Записал:")

	for i := range created {
		b.WriteString("\n• ")
		b.WriteString(formatExpense(created[i], today))
	}

	return b.String()
}

func formatUpdated(before, after models.Expense, today models.Date) string {
	var changes []string
	if before.Amount != after.Amount {
		changes = append(changes, "сумма "+before.Amount.String()+" → "+after.Amount.String())
	}

	if before.Category != after.Category {
		changes = append(changes, "категория "+before.Category.Label()+" → "+after.Category.Label())
	}

	if before.SpentAt != after.SpentAt {
		changes = append(changes, "дата "+humanDate(before.SpentAt, today)+" → "+humanDate(after.SpentAt, today))
	}

	if before.Comment != after.Comment {
		changes = append(changes, "«"+before.Comment+"» → «"+after.Comment+"»")
	}

	if len(changes) == 0 {
		return "Ничего не изменилось: " + formatExpense(after, today)
	}

	return "✏️ Исправил: " + strings.Join(changes, ", ") + "\nТеперь: " + formatExpense(after, today)
}

func formatList(label string, category models.Category, s models.Summary, items []models.Expense, today models.Date) string {
	subject := "Траты " + label
	if category != "" {
		subject = "Траты на «" + category.Label() + "» " + label
	}

	if s.Count == 0 {
		return subject + ": ничего не нашёл."
	}
	var b strings.Builder
	b.WriteString(subject + ": " + s.Total.String() + " (" + countWord(s.Count) + ")")

	if category == "" && len(s.ByCategory) > 1 {
		b.WriteString("\n\nПо категориям:")

		cats := make([]models.Category, 0, len(s.ByCategory))
		for c := range s.ByCategory {
			cats = append(cats, c)
		}

		slices.SortFunc(cats, func(x, y models.Category) int { return int(s.ByCategory[y] - s.ByCategory[x]) })

		for _, c := range cats {
			b.WriteString("\n• " + c.Label() + " — " + s.ByCategory[c].String())
		}
	}

	b.WriteString("\n\nПоследние:")

	for i := range items {
		b.WriteString("\n• " + formatExpense(items[i], today))
	}

	if s.Count > len(items) {
		b.WriteString("\n…и ещё " + countWord(s.Count-len(items)))
	}

	return b.String()
}

func countWord(n int) string {
	word := "трат"
	switch mod10, mod100 := n%10, n%100; {
	case mod10 == 1 && mod100 != 11:
		word = "трата"
	case mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14):
		word = "траты"
	}

	return itoa(n) + " " + word
}
