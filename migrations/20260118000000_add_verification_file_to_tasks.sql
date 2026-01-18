-- +goose Up
ALTER TABLE tasks ADD COLUMN verification_file TEXT;

-- +goose Down
ALTER TABLE tasks DROP COLUMN verification_file;
