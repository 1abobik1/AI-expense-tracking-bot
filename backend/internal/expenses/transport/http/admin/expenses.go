package admin

import (
	"net/http"

	"expensebot/internal/expenses/models"

	apphttp "expensebot/internal/app/http"
)

var scope = models.AdminScope()

func (h *Handlers) getCategories(w http.ResponseWriter, _ *http.Request) {
	out := make([]categoryOut, 0, len(models.Categories))
	for _, c := range models.Categories {
		out = append(out, categoryOut{Code: string(c), Label: c.Label()})
	}

	apphttp.WriteJSON(w, http.StatusOK, out)
}

func (h *Handlers) getExpenses(w http.ResponseWriter, r *http.Request) {
	params, err := mapQueryToParams(r)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	items, err := h.expensesUC.GetExpenses(r.Context(), scope, params)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	summary, err := h.expensesUC.GetSummary(r.Context(), scope, params)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	out := expensesOut{Items: make([]expenseOut, 0, len(items)), Total: mapSummaryToOut(*summary)}
	for i := range items {
		out.Items = append(out.Items, mapExpenseToOut(items[i]))
	}

	apphttp.WriteJSON(w, http.StatusOK, out)
}

func (h *Handlers) createExpense(w http.ResponseWriter, r *http.Request) {
	var in createExpenseIn
	if !apphttp.Decode(w, r, &in) {
		return
	}

	e, err := mapCreateIn(in)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	created, err := h.expensesUC.CreateExpense(r.Context(), scope, e)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusCreated, mapExpenseToOut(*created))
}

func (h *Handlers) getExpense(w http.ResponseWriter, r *http.Request) {
	id, ok := apphttp.PathID(w, r)
	if !ok {
		return
	}

	e, err := h.expensesUC.GetExpense(r.Context(), scope, id)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, mapExpenseToOut(*e))
}

func (h *Handlers) updateExpense(w http.ResponseWriter, r *http.Request) {
	id, ok := apphttp.PathID(w, r)
	if !ok {
		return
	}

	var in updateExpenseIn
	if !apphttp.Decode(w, r, &in) {
		return
	}

	upd, err := mapUpdateIn(in)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	e, err := h.expensesUC.UpdateExpense(r.Context(), scope, id, upd)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	apphttp.WriteJSON(w, http.StatusOK, mapExpenseToOut(*e))
}

func (h *Handlers) deleteExpense(w http.ResponseWriter, r *http.Request) {
	id, ok := apphttp.PathID(w, r)
	if !ok {
		return
	}

	if _, err := h.expensesUC.DeleteExpense(r.Context(), scope, id); err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
