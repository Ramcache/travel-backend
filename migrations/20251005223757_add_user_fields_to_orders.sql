-- +goose Up
ALTER TABLE orders
    ADD COLUMN name TEXT,
    ADD COLUMN date TEXT,
    ADD COLUMN price TEXT;

-- +goose Down
ALTER TABLE orders
    DROP COLUMN name,
    DROP COLUMN date,
    DROP COLUMN price;