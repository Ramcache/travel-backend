-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS avatar;
