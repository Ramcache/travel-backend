-- +goose Up
ALTER TABLE trips ADD COLUMN active boolean NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE trips DROP COLUMN active;
