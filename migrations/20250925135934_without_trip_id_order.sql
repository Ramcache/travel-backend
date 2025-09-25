-- +goose Up
ALTER TABLE orders
    ALTER COLUMN trip_id DROP NOT NULL;

-- +goose Down
ALTER TABLE orders
    ALTER COLUMN trip_id SET NOT NULL;
