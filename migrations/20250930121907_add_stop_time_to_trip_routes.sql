-- +goose Up
ALTER TABLE trip_routes
    ADD COLUMN stop_time VARCHAR(50);

-- +goose Down
ALTER TABLE trip_routes
    DROP COLUMN IF EXISTS stop_time;
