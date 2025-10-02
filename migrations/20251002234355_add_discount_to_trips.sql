-- +goose Up
ALTER TABLE trips
    ADD COLUMN discount_percent INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE trips DROP COLUMN discount_percent;