package counter_test

import (
	"context"
	"testing"

	"expensebot/internal/limits/repository/sqlite/counter"
	"expensebot/internal/testutil"
)

func TestCounters(t *testing.T) {
	ctx := context.Background()

	repo := counter.NewRepo(testutil.SQLite(t))
	for want := 1; want <= 3; want++ {
		if n, err := repo.Incr(ctx, "msg:1", "2026-10-15"); err != nil || n != want {
			t.Fatalf("Incr = %d, %v; want %d", n, err, want)
		}
	}

	if n, _ := repo.Incr(ctx, "msg:1", "2026-10-16"); n != 1 {
		t.Fatalf("new bucket starts at %d", n)
	}

	if err := repo.DeleteBefore(ctx, "2026-10-16"); err != nil {
		t.Fatal(err)
	}

	if n, _ := repo.Incr(ctx, "msg:1", "2026-10-15"); n != 1 {
		t.Fatalf("old bucket was not deleted: %d", n)
	}
}
