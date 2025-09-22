-- +goose Up
ALTER TABLE news
    DROP COLUMN IF EXISTS category;

-- +goose Down
ALTER TABLE news
    ADD COLUMN category TEXT;
