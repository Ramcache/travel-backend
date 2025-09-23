-- +goose Up
ALTER TABLE trips
    ADD COLUMN views_count INT NOT NULL DEFAULT 0,
    ADD COLUMN buys_count INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE trips
    DROP COLUMN views_count,
    DROP COLUMN buys_count;
