package testutil

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"expensebot/internal/app/sqlite/conn"
)

// SQLite — временная база с миграциями goose; тот же SQL, что у D1.
func SQLite(t *testing.T) *sql.DB {
	t.Helper()

	db, err := conn.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { db.Close() })

	if err := conn.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	return db
}
