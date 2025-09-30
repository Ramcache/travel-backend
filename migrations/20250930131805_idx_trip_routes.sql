-- +goose Up
CREATE INDEX IF NOT EXISTS idx_trip_routes_trip_position ON trip_routes (trip_id, position);

-- +goose Down
DROP INDEX IF EXISTS idx_trip_routes_trip_position;
