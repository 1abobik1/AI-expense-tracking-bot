-- +goose Up
-- Даты и время — TEXT в ISO: так их одинаково отдают SQLite и D1, а строки сортируются как время.
CREATE TABLE members (
    user_id    INTEGER PRIMARY KEY,
    username   TEXT NOT NULL DEFAULT '',
    first_name TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE expenses (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES members (user_id),
    amount     INTEGER NOT NULL CHECK (amount > 0), -- в копейках
    category   TEXT    NOT NULL CHECK (category IN ('food', 'transport', 'housing', 'entertainment', 'health', 'shopping', 'other')),
    spent_at   TEXT    NOT NULL,
    comment    TEXT    NOT NULL DEFAULT '',
    created_at TEXT    NOT NULL,
    updated_at TEXT    NOT NULL
);

CREATE INDEX expenses_user_spent_at_idx ON expenses (user_id, spent_at DESC, id DESC);
CREATE INDEX expenses_spent_at_idx ON expenses (spent_at DESC, id DESC);

-- Счётчики лимитов: сообщения за сутки, вызовы ИИ, попытки входа.
CREATE TABLE counters (
    key    TEXT    NOT NULL,
    bucket TEXT    NOT NULL,
    count  INTEGER NOT NULL,
    PRIMARY KEY (key, bucket)
);

CREATE INDEX counters_bucket_idx ON counters (bucket);

-- +goose Down
DROP TABLE counters;
DROP TABLE expenses;
DROP TABLE members;
