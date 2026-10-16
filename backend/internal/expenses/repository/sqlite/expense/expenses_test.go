package expense_test

import (
	"context"
	"testing"

	"expensebot/internal/app/apperr"
	"expensebot/internal/expenses/models"
	"expensebot/internal/expenses/repository/sqlite/expense"
	"expensebot/internal/members/repository/sqlite/member"
	"expensebot/internal/testutil"

	membersModels "expensebot/internal/members/models"
)

const (
	alice int64 = 1001
	bob   int64 = 2002
)

func setup(t *testing.T) (repo *expense.Repo, aliceTaxi, bobCoffee models.Expense) {
	t.Helper()
	db := testutil.SQLite(t)
	ctx := context.Background()

	members := member.NewRepo(db)
	for _, id := range []int64{alice, bob} {
		if err := members.UpsertMember(ctx, &membersModels.Member{UserID: id}); err != nil {
			t.Fatal(err)
		}
	}

	repo = expense.NewRepo(db)
	aliceTaxi = models.Expense{
		UserID: alice, Amount: 1200_00, Category: models.CategoryTransport,
		SpentAt: models.NewDate(2026, 10, 14), Comment: "такси",
	}

	bobCoffee = models.Expense{
		UserID: bob, Amount: 350_50, Category: models.CategoryFood,
		SpentAt: models.NewDate(2026, 10, 15), Comment: "кофе",
	}
	for _, e := range []*models.Expense{&aliceTaxi, &bobCoffee} {
		if err := repo.CreateExpense(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	return repo, aliceTaxi, bobCoffee
}

func TestExpenseCRUD(t *testing.T) {
	ctx := context.Background()

	repo, taxi, _ := setup(t)
	if taxi.ID == 0 || taxi.CreatedAt.IsZero() {
		t.Fatalf("CreateExpense did not fill ID/CreatedAt: %+v", taxi)
	}

	got, err := repo.GetExpense(ctx, models.UserScope(alice), taxi.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.Amount != taxi.Amount || got.SpentAt != taxi.SpentAt || got.Comment != "такси" || !got.CreatedAt.Equal(taxi.CreatedAt) {
		t.Fatalf("GetExpense = %+v, want %+v", got, taxi)
	}

	got.Amount = 1500_00
	if err := repo.UpdateExpense(ctx, models.UserScope(alice), got); err != nil {
		t.Fatal(err)
	}

	if got, _ = repo.GetExpense(ctx, models.AdminScope(), taxi.ID); got.Amount != 1500_00 {
		t.Fatalf("amount after update = %v", got.Amount)
	}

	if err := repo.DeleteExpense(ctx, models.UserScope(alice), taxi.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.GetExpense(ctx, models.AdminScope(), taxi.ID); !apperr.Is(err, apperr.CodeNotFound) {
		t.Fatalf("GetExpense after delete: %v", err)
	}
}

func TestUserScopeHidesOtherUsers(t *testing.T) {
	ctx := context.Background()
	repo, taxi, coffee := setup(t)
	asBob := models.UserScope(bob)

	list, err := repo.GetExpenses(ctx, asBob, models.GetExpensesParams{UserID: alice, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 1 || list[0].ID != coffee.ID {
		t.Fatalf("bob sees %+v, want only his coffee (params.UserID must be ignored)", list)
	}

	if _, err := repo.GetExpense(ctx, asBob, taxi.ID); !apperr.Is(err, apperr.CodeNotFound) {
		t.Fatalf("bob GetExpense alice's: %v", err)
	}

	hijacked := taxi

	hijacked.Amount = 1
	if err := repo.UpdateExpense(ctx, asBob, &hijacked); !apperr.Is(err, apperr.CodeNotFound) {
		t.Fatalf("bob UpdateExpense alice's: %v", err)
	}

	if err := repo.DeleteExpense(ctx, asBob, taxi.ID); !apperr.Is(err, apperr.CodeNotFound) {
		t.Fatalf("bob DeleteExpense alice's: %v", err)
	}

	if got, _ := repo.GetExpense(ctx, models.UserScope(alice), taxi.ID); got.Amount != taxi.Amount {
		t.Fatalf("alice's expense changed: %+v", got)
	}

	if s, _ := repo.GetSummary(ctx, asBob, models.GetExpensesParams{}); s.Count != 1 || s.Total != coffee.Amount {
		t.Fatalf("bob summary = %+v", s)
	}

	if list, _ := repo.GetExpenses(ctx, models.Scope{}, models.GetExpensesParams{Limit: 100}); len(list) != 0 {
		t.Fatalf("zero scope sees %d expenses", len(list))
	}
}

func TestFiltersAndSummary(t *testing.T) {
	ctx := context.Background()
	repo, taxi, coffee := setup(t)

	cases := []struct {
		name   string
		params models.GetExpensesParams
		want   []int64
	}{
		{"all, newest first", models.GetExpensesParams{}, []int64{coffee.ID, taxi.ID}},
		{"member", models.GetExpensesParams{UserID: alice}, []int64{taxi.ID}},
		{"category", models.GetExpensesParams{Category: models.CategoryFood}, []int64{coffee.ID}},
		{"period", models.GetExpensesParams{From: models.NewDate(2026, 10, 14), To: models.NewDate(2026, 10, 14)}, []int64{taxi.ID}},
		{"offset", models.GetExpensesParams{Offset: 1}, []int64{taxi.ID}},
	}
	for _, c := range cases {
		c.params.Limit = 100

		list, err := repo.GetExpenses(ctx, models.AdminScope(), c.params)
		if err != nil {
			t.Fatal(err)
		}

		if len(list) != len(c.want) || list[0].ID != c.want[0] {
			t.Errorf("%s: got %d items, want ids %v", c.name, len(list), c.want)
		}
	}

	s, err := repo.GetSummary(ctx, models.AdminScope(), models.GetExpensesParams{})
	if err != nil {
		t.Fatal(err)
	}

	if s.Count != 2 || s.Total != taxi.Amount+coffee.Amount || s.ByCategory[models.CategoryFood] != coffee.Amount {
		t.Fatalf("summary = %+v", s)
	}
}
