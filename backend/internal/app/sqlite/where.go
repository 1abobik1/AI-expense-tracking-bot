package sqlite

import (
	"database/sql"
	"strings"
)

// Where собирает условия WHERE вместе с аргументами.
type Where struct {
	conds []string
	args  []any
}

func (w *Where) Add(cond string, arg any) {
	w.conds = append(w.conds, cond)
	w.args = append(w.args, arg)
}

func (w *Where) SQL() string {
	if len(w.conds) == 0 {
		return ""
	}

	return " WHERE " + strings.Join(w.conds, " AND ")
}

func (w *Where) Args() []any { return w.args }

// AffectedOne возвращает notFound, если запрос не затронул ни одной строки.
func AffectedOne(res sql.Result, err, notFound error) error {
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return notFound
	}

	return nil
}
