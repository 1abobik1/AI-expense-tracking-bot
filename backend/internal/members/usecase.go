package members

import (
	"context"

	"expensebot/internal/members/models"
)

type MembersUseCase interface {
	TouchMember(ctx context.Context, m models.Member) (*models.Member, error)
	GetMembers(ctx context.Context) ([]models.Member, error)
	UpdateMember(ctx context.Context, userID int64, in models.UpdateMemberIn) (*models.Member, error)
}
