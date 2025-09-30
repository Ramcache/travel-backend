-- +goose Up
ALTER TABLE hotels ADD COLUMN transfer TEXT DEFAULT 'не указано';
ALTER TABLE hotels ADD COLUMN capacity TEXT DEFAULT 'не указано';

-- +goose Down
ALTER TABLE hotels DROP COLUMN transfer;
ALTER TABLE hotels DROP COLUMN capacity;
