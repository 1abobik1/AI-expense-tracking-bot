package usecase

import (
	"strings"
	"testing"

	"expensebot/internal/expenses/models"
)

func TestParseIntent(t *testing.T) {
	for _, text := range []string{
		`{"action":"ADD"}`,
		"```json\n{\"action\":\"add\"}\n```",
		`Конечно! {"action":"add"} Готово.`,
	} {
		if in, ok := parseIntent(text); !ok || in.Action != "add" {
			t.Errorf("parseIntent(%q) = %+v, %v", text, in, ok)
		}
	}

	for _, text := range []string{"", "нет json", `{"action":""}`, `{broken`} {
		if _, ok := parseIntent(text); ok {
			t.Errorf("parseIntent(%q) must fail", text)
		}
	}
}

func TestResolvePeriod(t *testing.T) {
	today := models.NewDate(2026, 10, 15) // четверг

	cases := []struct {
		in       Intent
		from, to string
	}{
		{Intent{Period: "today"}, "2026-10-15", "2026-10-15"},
		{Intent{Period: "yesterday"}, "2026-10-14", "2026-10-14"},
		{Intent{Period: "this_week"}, "2026-10-12", "2026-10-15"},
		{Intent{Period: "last_7_days"}, "2026-10-09", "2026-10-15"},
		{Intent{Period: "this_month"}, "2026-10-01", "2026-10-15"},
		{Intent{Period: "last_month"}, "2026-09-01", "2026-09-30"},
		{Intent{Period: "this_year"}, "2026-01-01", "2026-10-15"},
		{Intent{Period: "all"}, "", ""},
		{Intent{Period: "custom", From: "2026-10-01", To: "2026-10-05"}, "2026-10-01", "2026-10-05"},
		{Intent{Period: "непонятно"}, "2026-10-01", "2026-10-15"},
	}
	for _, c := range cases {
		from, to, _ := resolvePeriod(c.in, today)
		if from.String() != c.from || to.String() != c.to {
			t.Errorf("%+v: %s..%s, want %s..%s", c.in, from, to, c.from, c.to)
		}
	}
}

func TestMatchesText(t *testing.T) {
	e := models.Expense{Comment: "Такси домой", Category: models.CategoryTransport}
	for text, want := range map[string]bool{"такси": true, "таксист": true, "транспорт": true, "кофе": false} {
		if got := matchesText(e, text); got != want {
			t.Errorf("matchesText(%q) = %v", text, got)
		}
	}
}

func TestPromptHasCalendar(t *testing.T) {
	p := systemPrompt(models.NewDate(2026, 10, 15))
	for _, part := range []string{"2026-10-15 четверг (сегодня)", "2026-10-14 среда (вчера)", "food (еда)"} {
		if !strings.Contains(p, part) {
			t.Errorf("prompt has no %q", part)
		}
	}
}

func TestCountWord(t *testing.T) {
	for n, want := range map[int]string{1: "1 трата", 3: "3 траты", 5: "5 трат", 11: "11 трат", 21: "21 трата", 112: "112 трат"} {
		if got := countWord(n); got != want {
			t.Errorf("countWord(%d) = %q", n, got)
		}
	}
}
