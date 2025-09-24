-- +goose Up
ALTER TABLE orders ADD COLUMN is_read BOOLEAN DEFAULT false;

-- +goose Down
ALTER TABLE orders DROP COLUMN is_read;
