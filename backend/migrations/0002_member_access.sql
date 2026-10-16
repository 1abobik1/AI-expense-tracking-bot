-- +goose Up
ALTER TABLE members ADD COLUMN name TEXT NOT NULL DEFAULT '';
ALTER TABLE members ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));

-- +goose Down
ALTER TABLE members DROP COLUMN status;
ALTER TABLE members DROP COLUMN name;
