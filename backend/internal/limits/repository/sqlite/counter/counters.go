package counter

import (
	"context"
	"database/sql"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Incr(ctx context.Context, key, bucket string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, incrQuery, key, bucket).Scan(&n)

	return n, err
}

func (r *Repo) DeleteBefore(ctx context.Context, bucket string) error {
	_, err := r.db.ExecContext(ctx, deleteBeforeQuery, bucket)
	return err
}
