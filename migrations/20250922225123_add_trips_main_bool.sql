-- +goose Up
ALTER TABLE trips ADD COLUMN main BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE trips DROP COLUMN main;
