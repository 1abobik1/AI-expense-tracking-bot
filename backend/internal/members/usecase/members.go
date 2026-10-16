package usecase

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"expensebot/internal/members"
	"expensebot/internal/members/models"
)

const approvedText = "Глава семьи открыл доступ. Пиши траты обычным текстом, например: «кофе 350»."

type MembersUseCase struct {
	repo     members.MembersRepo
	notifier members.Notifier
}

func NewMembersUC(repo members.MembersRepo, notifier members.Notifier) *MembersUseCase {
	return &MembersUseCase{repo: repo, notifier: notifier}
}

func (uc *MembersUseCase) TouchMember(ctx context.Context, m models.Member) (*models.Member, error) {
	if err := uc.repo.UpsertMember(ctx, &m); err != nil {
		return nil, fmt.Errorf("repo.UpsertMember: %w", err)
	}

	return uc.repo.GetMember(ctx, m.UserID)
}

func (uc *MembersUseCase) GetMembers(ctx context.Context) ([]models.Member, error) {
	return uc.repo.GetMembers(ctx)
}

func (uc *MembersUseCase) UpdateMember(ctx context.Context, userID int64, in models.UpdateMemberIn) (*models.Member, error) {
	m, err := uc.repo.GetMember(ctx, userID)
	if err != nil {
		return nil, err
	}

	wasApproved := m.Status == models.StatusApproved
	if in.Status != nil {
		if !in.Status.Valid() {
			return nil, members.ErrUnknownStatus()
		}

		m.Status = *in.Status
	}

	if in.Name != nil {
		m.Name = strings.TrimSpace(*in.Name)
		if utf8.RuneCountInString(m.Name) > models.MaxNameLength {
			return nil, members.ErrNameTooLong()
		}
	}

	if err := uc.repo.UpdateMember(ctx, m); err != nil {
		return nil, err
	}

	if !wasApproved && m.Status == models.StatusApproved && uc.notifier != nil {
		_ = uc.notifier.SendMessage(ctx, m.UserID, approvedText) // доступ уже выдан, сообщение — бонус
	}

	return m, nil
}
