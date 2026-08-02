-- +goose Up
ALTER TABLE transactions DROP COLUMN source_id;
ALTER TABLE transactions ADD COLUMN source_name TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE transactions DROP COLUMN source_name;
ALTER TABLE transactions ADD COLUMN source_id BIGINT;
