package admin

import (
	"net/http"

	"expensebot/internal/app/apperr"
	"expensebot/internal/expenses/models"

	apphttp "expensebot/internal/app/http"
)

func mapExpenseToOut(e models.Expense) expenseOut {
	return expenseOut{
		ID:        e.ID,
		UserID:    e.UserID,
		Amount:    e.Amount.Rubles(),
		Category:  string(e.Category),
		SpentAt:   e.SpentAt.String(),
		Comment:   e.Comment,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func mapSummaryToOut(s models.Summary) totalOut {
	by := make(map[string]float64, len(s.ByCategory))
	for c, m := range s.ByCategory {
		by[string(c)] = m.Rubles()
	}

	return totalOut{Count: s.Count, Amount: s.Total.Rubles(), ByCategory: by}
}

func mapCreateIn(in createExpenseIn) (models.Expense, error) {
	spentAt, err := parseDate("spent_at", in.SpentAt)
	if err != nil {
		return models.Expense{}, err
	}

	return models.Expense{
		UserID:   in.UserID,
		Amount:   models.MoneyFromRubles(in.Amount),
		Category: models.Category(in.Category),
		SpentAt:  spentAt,
		Comment:  in.Comment,
	}, nil
}

func mapUpdateIn(in updateExpenseIn) (models.UpdateExpenseIn, error) {
	var out models.UpdateExpenseIn
	if in.Amount != nil {
		m := models.MoneyFromRubles(*in.Amount)
		out.Amount = &m
	}

	if in.Category != nil {
		c := models.Category(*in.Category)
		out.Category = &c
	}

	if in.SpentAt != nil {
		d, err := parseDate("spent_at", *in.SpentAt)
		if err != nil {
			return out, err
		}

		out.SpentAt = &d
	}

	out.Comment = in.Comment

	return out, nil
}

func mapQueryToParams(r *http.Request) (models.GetExpensesParams, error) {
	var p models.GetExpensesParams
	q := r.URL.Query()

	var err error
	if p.UserID, err = apphttp.QueryInt(r, "user_id"); err != nil {
		return p, err
	}

	if v := q.Get("from"); v != "" {
		if p.From, err = parseDate("from", v); err != nil {
			return p, err
		}
	}

	if v := q.Get("to"); v != "" {
		if p.To, err = parseDate("to", v); err != nil {
			return p, err
		}
	}

	if v := q.Get("category"); v != "" {
		if p.Category = models.Category(v); !p.Category.Valid() {
			return p, apperr.ErrValidation("category", "неизвестная категория")
		}
	}

	limit, err := apphttp.QueryInt(r, "limit")
	if err != nil {
		return p, err
	}

	offset, err := apphttp.QueryInt(r, "offset")
	p.Limit, p.Offset = int(limit), int(offset)

	return p, err
}

func parseDate(field, s string) (models.Date, error) {
	d, err := models.ParseDate(s)
	if err != nil {
		return d, apperr.ErrValidation(field, "дата должна быть в формате ГГГГ-ММ-ДД")
	}

	return d, nil
}
