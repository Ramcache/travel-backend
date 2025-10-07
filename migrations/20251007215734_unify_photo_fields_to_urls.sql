-- +goose Up
ALTER TABLE trips
    DROP COLUMN IF EXISTS photo_url,
    ADD COLUMN urls TEXT[] DEFAULT '{}' NOT NULL;

ALTER TABLE hotels
    DROP COLUMN IF EXISTS photo_url,
    ADD COLUMN urls TEXT[] DEFAULT '{}' NOT NULL;

ALTER TABLE news
    DROP COLUMN IF EXISTS preview_url,
    ADD COLUMN urls TEXT[] DEFAULT '{}' NOT NULL;

-- +goose Down
ALTER TABLE trips
    DROP COLUMN IF EXISTS urls,
    ADD COLUMN photo_url TEXT;

ALTER TABLE hotels
    DROP COLUMN IF EXISTS urls,
    ADD COLUMN photo_url TEXT;

ALTER TABLE news
    DROP COLUMN IF EXISTS urls,
    ADD COLUMN preview_url TEXT;
