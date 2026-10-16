package members

import (
	"context"

	"expensebot/internal/members/models"
)

type MembersRepo interface {
	UpsertMember(ctx context.Context, m *models.Member) error
	UpdateMember(ctx context.Context, m *models.Member) error
	GetMember(ctx context.Context, userID int64) (*models.Member, error)
	GetMembers(ctx context.Context) ([]models.Member, error)
}
