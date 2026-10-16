package member_test

import (
	"context"
	"testing"

	"expensebot/internal/app/apperr"
	"expensebot/internal/members/models"
	"expensebot/internal/members/repository/sqlite/member"
	"expensebot/internal/testutil"
)

func TestMembers(t *testing.T) {
	ctx := context.Background()

	repo := member.NewRepo(testutil.SQLite(t))
	if err := repo.UpsertMember(ctx, &models.Member{UserID: 1, Username: "alice", FirstName: "Алиса"}); err != nil {
		t.Fatal(err)
	}

	m, err := repo.GetMember(ctx, 1)
	if err != nil || m.FirstName != "Алиса" || m.Status != models.StatusPending {
		t.Fatalf("member = %+v, err = %v", m, err)
	}

	m.Status, m.Name = models.StatusApproved, "Мама"
	if err := repo.UpdateMember(ctx, m); err != nil {
		t.Fatal(err)
	}
	// Сообщение боту обновляет данные из Telegram, но не сбрасывает статус и имя от главы семьи.
	if err := repo.UpsertMember(ctx, &models.Member{UserID: 1, Username: "alice2", FirstName: "Алиса"}); err != nil {
		t.Fatal(err)
	}

	if m, _ = repo.GetMember(ctx, 1); m.Status != models.StatusApproved || m.Name != "Мама" || m.Username != "alice2" {
		t.Fatalf("after upsert: %+v", m)
	}

	if err := repo.UpdateMember(ctx, &models.Member{UserID: 2, Status: models.StatusApproved}); !apperr.Is(err, apperr.CodeNotFound) {
		t.Fatalf("update unknown member: %v", err)
	}

	if list, _ := repo.GetMembers(ctx); len(list) != 1 {
		t.Fatalf("members = %+v", list)
	}
}
