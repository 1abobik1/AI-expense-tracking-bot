package usecase_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"expensebot/internal/agent"
	"expensebot/internal/agent/usecase"
	"expensebot/internal/app/clock"
	"expensebot/internal/expenses/models"
	"expensebot/internal/testutil"

	expenseRepo "expensebot/internal/expenses/repository/sqlite/expense"
	expensesUsecase "expensebot/internal/expenses/usecase"
	counterRepo "expensebot/internal/limits/repository/sqlite/counter"
	limitsUsecase "expensebot/internal/limits/usecase"
	membersModels "expensebot/internal/members/models"
	memberRepo "expensebot/internal/members/repository/sqlite/member"
)

const (
	alice int64 = 1001
	bob   int64 = 2002
)

// scriptedLLM отвечает заранее заданным JSON на каждое сообщение.
type scriptedLLM struct {
	answers map[string]string
	calls   int
}

func (l *scriptedLLM) Complete(_ context.Context, req agent.LLMRequest) (string, error) {
	l.calls++
	return l.answers[req.User], nil
}

type env struct {
	agent    *usecase.AgentUseCase
	llm      *scriptedLLM
	expenses *expenseRepo.Repo
}

// «Сейчас» в тестах — четверг, 15 октября 2026.
var now = time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)

func newEnv(t *testing.T, limits usecase.Limits, answers map[string]string) *env {
	t.Helper()
	db := testutil.SQLite(t)
	ctx := context.Background()

	members := memberRepo.NewRepo(db)
	for _, id := range []int64{alice, bob} {
		if err := members.UpsertMember(ctx, &membersModels.Member{UserID: id}); err != nil {
			t.Fatal(err)
		}
	}

	loc, _ := time.LoadLocation("Europe/Moscow")
	clk := clock.Clock{Location: loc, Now: func() time.Time { return now }}
	llm := &scriptedLLM{answers: answers}
	expenses := expenseRepo.NewRepo(db)
	expensesUC := expensesUsecase.NewExpensesUC(expenses, members, clk)
	limiter := limitsUsecase.NewLimiter(counterRepo.NewRepo(db), clk)
	a := usecase.NewAgentUC(expensesUC, llm, limiter, limits, slog.New(slog.NewTextHandler(io.Discard, nil)))

	return &env{agent: a, llm: llm, expenses: expenses}
}

func (e *env) say(t *testing.T, user int64, text string) string {
	t.Helper()

	reply, err := e.agent.Handle(context.Background(), user, text)
	if err != nil {
		t.Fatalf("Handle(%q): %v", text, err)
	}

	return reply
}

func (e *env) expensesOf(t *testing.T, user int64) []models.Expense {
	t.Helper()

	list, err := e.expenses.GetExpenses(context.Background(), models.UserScope(user), models.GetExpensesParams{Limit: 100})
	if err != nil {
		t.Fatal(err)
	}

	return list
}

func wantContains(t *testing.T, reply string, parts ...string) {
	t.Helper()

	for _, p := range parts {
		if !strings.Contains(reply, p) {
			t.Errorf("reply %q does not contain %q", reply, p)
		}
	}
}

var script = map[string]string{
	"кофе 350":         `{"action":"add","items":[{"amount":350,"category":"food","date":"2026-10-15","comment":"кофе"}]}`,
	"такси 1200 вчера": "```json\n" + `{"action":"add","items":[{"amount":1200,"category":"transport","date":"2026-10-14","comment":"такси"}]}` + "\n```",
	"сколько на еду":   `{"action":"list","period":"this_month","category":"food"}`,
	"покажи за неделю": `Вот: {"action":"list","period":"this_week"}`,
	"последнее 450":    `{"action":"update","target":{"last":true,"amount":350},"changes":{"amount":450}}`,
	"удали такси":      `{"action":"delete","target":{"text":"такси"}}`,
	"привет":           `{"action":"help"}`,
	"мусор":            `не знаю`,
	"завтра 100":       `{"action":"add","items":[{"amount":100,"category":"food","date":"2026-10-20","comment":"обед"}]}`,
	// Попытка подсунуть чужой user_id: в Intent такого поля нет, оно игнорируется.
	"injection": `{"action":"add","user_id":1001,"items":[{"amount":1,"category":"food","comment":"взлом","user_id":1001}]}`,
}

func TestCRUDConversation(t *testing.T) {
	e := newEnv(t, usecase.Limits{}, script)

	wantContains(t, e.say(t, alice, "кофе 350"), "Записал", "кофе", "350 ₽", "еда", "сегодня")
	wantContains(t, e.say(t, alice, "такси 1200 вчера"), "такси", "1 200 ₽", "вчера")

	if got := len(e.expensesOf(t, alice)); got != 2 {
		t.Fatalf("alice has %d expenses, want 2", got)
	}

	wantContains(t, e.say(t, alice, "сколько на еду"), "350 ₽", "1 трата")
	wantContains(t, e.say(t, alice, "покажи за неделю"), "1 550 ₽", "2 траты", "транспорт — 1 200 ₽")

	wantContains(t, e.say(t, alice, "последнее 450"), "350 ₽ → 450 ₽")
	wantContains(t, e.say(t, alice, "удали такси"), "Удалил", "такси")

	list := e.expensesOf(t, alice)
	if len(list) != 1 || list[0].Amount != 450_00 {
		t.Fatalf("after update and delete: %+v", list)
	}
}

func TestOtherUsersExpensesAreInvisible(t *testing.T) {
	e := newEnv(t, usecase.Limits{}, script)
	e.say(t, alice, "такси 1200 вчера")

	wantContains(t, e.say(t, bob, "покажи за неделю"), "ничего не нашёл")
	wantContains(t, e.say(t, bob, "удали такси"), "Не нашёл")
	wantContains(t, e.say(t, bob, "последнее 450"), "Не нашёл")

	if list := e.expensesOf(t, alice); len(list) != 1 || list[0].Amount != 1200_00 {
		t.Fatalf("alice's expense was touched: %+v", list)
	}

	e.say(t, bob, "injection")

	if len(e.expensesOf(t, bob)) != 1 || len(e.expensesOf(t, alice)) != 1 {
		t.Fatal("injected user_id must be ignored: the expense belongs to the author")
	}
}

func TestAmbiguousDeleteAsksToClarify(t *testing.T) {
	e := newEnv(t, usecase.Limits{}, script)
	e.say(t, alice, "такси 1200 вчера")
	e.say(t, alice, "такси 1200 вчера")
	wantContains(t, e.say(t, alice, "удали такси"), "несколько")

	if len(e.expensesOf(t, alice)) != 2 {
		t.Fatal("nothing must be deleted when the target is ambiguous")
	}
}

func TestUnclearAndInvalid(t *testing.T) {
	e := newEnv(t, usecase.Limits{}, script)
	wantContains(t, e.say(t, alice, "привет"), "кофе 350")
	wantContains(t, e.say(t, alice, "мусор"), "кофе 350")
	wantContains(t, e.say(t, alice, "завтра 100"), "в будущем")

	if len(e.expensesOf(t, alice)) != 0 {
		t.Fatal("invalid expense was saved")
	}
}

func TestLimits(t *testing.T) {
	e := newEnv(t, usecase.Limits{MessagesPerUserPerDay: 2, LLMCallsPerDay: 3, MaxMessageLength: 20}, script)

	wantContains(t, e.say(t, alice, strings.Repeat("я", 21)), "длинное")
	e.say(t, alice, "кофе 350")
	e.say(t, alice, "кофе 350")
	wantContains(t, e.say(t, alice, "кофе 350"), "лимит сообщений")

	e.say(t, bob, "кофе 350")
	wantContains(t, e.say(t, bob, "кофе 350"), "ИИ на сегодня")

	if e.llm.calls != 3 {
		t.Fatalf("llm calls = %d, want 3: limited messages must not reach the model", e.llm.calls)
	}
}
