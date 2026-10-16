package models

import (
	"testing"

	"expensebot/internal/app/apperr"
)

func TestMoney(t *testing.T) {
	for m, want := range map[Money]string{
		350_00: "350 ₽", 1200_00: "1 200 ₽", 99_90: "99,90 ₽", 1_234_567_05: "1 234 567,05 ₽", -5_00: "-5 ₽",
	} {
		if got := m.String(); got != want {
			t.Errorf("Money(%d) = %q, want %q", m, got, want)
		}
	}

	if MoneyFromRubles(0.1+0.2) != 30 || MoneyFromRubles(1500) != 1500_00 {
		t.Error("MoneyFromRubles rounding")
	}
}

func TestParseCategory(t *testing.T) {
	cases := []struct {
		in   string
		want Category
	}{{"food", CategoryFood}, {"Еда", CategoryFood}, {"жилье", CategoryHousing}, {" жильё ", CategoryHousing}}
	for _, c := range cases {
		if got, ok := ParseCategory(c.in); !ok || got != c.want {
			t.Errorf("ParseCategory(%q) = %q, %v", c.in, got, ok)
		}
	}

	if _, ok := ParseCategory("кофе"); ok {
		t.Error("unknown category accepted")
	}
}

func TestExpenseValidate(t *testing.T) {
	valid := Expense{UserID: 1, Amount: 100, Category: CategoryFood, SpentAt: NewDate(2026, 10, 15)}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}

	broken := map[string]func(*Expense){
		"user_id":  func(e *Expense) { e.UserID = 0 },
		"amount":   func(e *Expense) { e.Amount = 0 },
		"category": func(e *Expense) { e.Category = "кофе" },
		"spent_at": func(e *Expense) { e.SpentAt = Date{} },
	}
	for field, breakIt := range broken {
		e := valid
		breakIt(&e)

		if err, ok := apperr.As(e.Validate()); !ok || err.Field != field {
			t.Errorf("%s: err = %v", field, err)
		}
	}
}
