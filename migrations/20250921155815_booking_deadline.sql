-- +goose Up
ALTER TABLE trips ADD COLUMN booking_deadline TIMESTAMP;

-- +goose Down
ALTER TABLE trips DROP COLUMN booking_deadline;
