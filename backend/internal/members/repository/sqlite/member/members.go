package member

import (
	"context"
	"database/sql"
	"errors"

	"expensebot/internal/app/sqlite"
	"expensebot/internal/members"
	"expensebot/internal/members/models"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) UpsertMember(ctx context.Context, m *models.Member) error {
	now := sqlite.FormatTime(sqlite.Now())
	_, err := r.db.ExecContext(ctx, upsertQuery, m.UserID, m.Username, m.FirstName, now, now)

	return err
}

func (r *Repo) UpdateMember(ctx context.Context, m *models.Member) error {
	res, err := r.db.ExecContext(ctx, updateQuery, m.Name, string(m.Status), sqlite.FormatTime(sqlite.Now()), m.UserID)
	return sqlite.AffectedOne(res, err, members.ErrMemberNotFound())
}

func (r *Repo) GetMember(ctx context.Context, userID int64) (*models.Member, error) {
	var ent memberEntity

	err := r.db.QueryRowContext(ctx, getQuery, userID).Scan(ent.fields()...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, members.ErrMemberNotFound()
	}

	if err != nil {
		return nil, err
	}

	m, err := entityToModel(ent)

	return &m, err
}

func (r *Repo) GetMembers(ctx context.Context) ([]models.Member, error) {
	rows, err := r.db.QueryContext(ctx, listQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Member
	for rows.Next() {
		var ent memberEntity
		if err := rows.Scan(ent.fields()...); err != nil {
			return nil, err
		}

		m, err := entityToModel(ent)
		if err != nil {
			return nil, err
		}

		out = append(out, m)
	}

	return out, rows.Err()
}
