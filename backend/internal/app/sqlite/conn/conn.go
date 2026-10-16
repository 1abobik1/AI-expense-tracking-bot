// Package conn — файл SQLite и миграции goose; только сервер, драйвер не собирается под wasm.
package conn

import (
	"context"
	"database/sql"
	"fmt"

	"expensebot/migrations"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// Open открывает файл БД; foreign keys в SQLite по умолчанию выключены.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1) // SQLite пишет в один поток; так не бывает SQLITE_BUSY

	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations.FS)
	if err != nil {
		return fmt.Errorf("goose.NewProvider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
