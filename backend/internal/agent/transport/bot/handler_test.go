package bot

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"expensebot/internal/agent/usecase"
	"expensebot/internal/app/telegram"
	"expensebot/internal/testutil"

	membersModels "expensebot/internal/members/models"
	memberRepo "expensebot/internal/members/repository/sqlite/member"
	membersUsecase "expensebot/internal/members/usecase"
)

type recorder struct{ sent []string }

func (r *recorder) SendMessage(_ context.Context, _ int64, text string) error {
	r.sent = append(r.sent, text)
	return nil
}
func (r *recorder) SendTyping(context.Context, int64) error { return nil }

func update(chatType, text string, isBot bool) telegram.Update {
	return telegram.Update{Message: &telegram.Message{
		From: &telegram.User{ID: 7, IsBot: isBot}, Chat: telegram.Chat{ID: 7, Type: chatType}, Text: text,
	}}
}

// Проверяются ветки до агента, поэтому агент — nil: если до него дойдёт, тест упадёт.
func TestHandleUpdateGuards(t *testing.T) {
	repo := memberRepo.NewRepo(testutil.SQLite(t))
	rec := &recorder{}
	h := NewHandler(nil, membersUsecase.NewMembersUC(repo, nil), rec, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx := context.Background()

	h.HandleUpdate(ctx, update("group", "кофе 350", false))
	h.HandleUpdate(ctx, update("private", "кофе 350", true))
	h.HandleUpdate(ctx, telegram.Update{})

	if len(rec.sent) != 0 {
		t.Fatalf("groups, bots and empty updates must be ignored, sent %q", rec.sent)
	}

	h.HandleUpdate(ctx, update("private", "кофе 350", false)) // новый — ждёт одобрения

	setStatus := func(s membersModels.Status) {
		m, _ := repo.GetMember(ctx, 7)
		m.Status = s
		repo.UpdateMember(ctx, m)
	}
	setStatus(membersModels.StatusRejected)
	h.HandleUpdate(ctx, update("private", "кофе 350", false))
	setStatus(membersModels.StatusApproved)
	h.HandleUpdate(ctx, update("private", "/start", false))
	h.HandleUpdate(ctx, update("private", "", false))

	want := []string{pendingText, rejectedText, usecase.HelpText, textOnly}
	if len(rec.sent) != len(want) {
		t.Fatalf("sent %q, want %q", rec.sent, want)
	}

	for i := range want {
		if rec.sent[i] != want[i] {
			t.Fatalf("sent %q, want %q", rec.sent, want)
		}
	}
}
