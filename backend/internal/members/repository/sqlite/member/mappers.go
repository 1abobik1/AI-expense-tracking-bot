package member

import (
	"fmt"

	"expensebot/internal/app/sqlite"
	"expensebot/internal/members/models"
)

func entityToModel(e memberEntity) (models.Member, error) {
	createdAt, err := sqlite.ParseTime(e.CreatedAt)
	if err != nil {
		return models.Member{}, fmt.Errorf("created_at %q: %w", e.CreatedAt, err)
	}

	return models.Member{
		UserID:    e.UserID,
		Username:  e.Username,
		FirstName: e.FirstName,
		Name:      e.Name,
		Status:    models.Status(e.Status),
		CreatedAt: createdAt,
	}, nil
}
