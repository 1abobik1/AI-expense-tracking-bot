package expense

import (
	"context"
	"database/sql"
	"errors"

	"expensebot/internal/app/sqlite"
	"expensebot/internal/expenses"
	"expensebot/internal/expenses/models"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) CreateExpense(ctx context.Context, e *models.Expense) error {
	now := sqlite.Now()

	err := r.db.QueryRowContext(ctx, insertQuery, e.UserID, int64(e.Amount), string(e.Category),
		e.SpentAt.String(), e.Comment, sqlite.FormatTime(now), sqlite.FormatTime(now)).Scan(&e.ID)
	if err != nil {
		return err
	}

	e.CreatedAt, e.UpdatedAt = now, now

	return nil
}

func (r *Repo) GetExpense(ctx context.Context, scope models.Scope, id int64) (*models.Expense, error) {
	w := whereScope(scope)
	w.Add("id = ?", id)
	var ent expenseEntity

	err := r.db.QueryRowContext(ctx, selectQuery+w.SQL(), w.Args()...).Scan(ent.fields()...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, expenses.ErrExpenseNotFound()
	}

	if err != nil {
		return nil, err
	}

	e, err := entityToModel(ent)

	return &e, err
}

func (r *Repo) GetExpenses(ctx context.Context, scope models.Scope, params models.GetExpensesParams) ([]models.Expense, error) {
	w := whereParams(scope, params)

	rows, err := r.db.QueryContext(ctx, selectQuery+w.SQL()+orderPage, append(w.Args(), params.Limit, params.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Expense
	for rows.Next() {
		var ent expenseEntity
		if err := rows.Scan(ent.fields()...); err != nil {
			return nil, err
		}

		e, err := entityToModel(ent)
		if err != nil {
			return nil, err
		}

		out = append(out, e)
	}

	return out, rows.Err()
}

func (r *Repo) GetSummary(ctx context.Context, scope models.Scope, params models.GetExpensesParams) (*models.Summary, error) {
	w := whereParams(scope, params)

	rows, err := r.db.QueryContext(ctx, summaryQuery+w.SQL()+" GROUP BY category", w.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s := &models.Summary{ByCategory: map[models.Category]models.Money{}}
	for rows.Next() {
		var (
			category string
			count    int
			total    int64
		)
		if err := rows.Scan(&category, &count, &total); err != nil {
			return nil, err
		}

		s.Count += count
		s.Total += models.Money(total)
		s.ByCategory[models.Category(category)] = models.Money(total)
	}

	return s, rows.Err()
}

func (r *Repo) UpdateExpense(ctx context.Context, scope models.Scope, e *models.Expense) error {
	now := sqlite.Now()
	w := whereScope(scope)
	w.Add("id = ?", e.ID)
	args := append([]any{int64(e.Amount), string(e.Category), e.SpentAt.String(), e.Comment, sqlite.FormatTime(now)}, w.Args()...)

	res, err := r.db.ExecContext(ctx, updateQuery+w.SQL(), args...)
	if err := sqlite.AffectedOne(res, err, expenses.ErrExpenseNotFound()); err != nil {
		return err
	}

	e.UpdatedAt = now

	return nil
}

func (r *Repo) DeleteExpense(ctx context.Context, scope models.Scope, id int64) error {
	w := whereScope(scope)
	w.Add("id = ?", id)
	res, err := r.db.ExecContext(ctx, deleteQuery+w.SQL(), w.Args()...)

	return sqlite.AffectedOne(res, err, expenses.ErrExpenseNotFound())
}

// whereScope — единственное место, где Scope превращается в SQL; с него начинается каждый запрос.
func whereScope(scope models.Scope) *sqlite.Where {
	w := &sqlite.Where{}
	if !scope.IsAdmin() {
		w.Add("user_id = ?", scope.UserID())
	}

	return w
}

func whereParams(scope models.Scope, p models.GetExpensesParams) *sqlite.Where {
	w := whereScope(scope)
	if scope.IsAdmin() && p.UserID != 0 {
		w.Add("user_id = ?", p.UserID)
	}

	if !p.From.IsZero() {
		w.Add("spent_at >= ?", p.From.String())
	}

	if !p.To.IsZero() {
		w.Add("spent_at <= ?", p.To.String())
	}

	if p.Category != "" {
		w.Add("category = ?", string(p.Category))
	}

	return w
}
