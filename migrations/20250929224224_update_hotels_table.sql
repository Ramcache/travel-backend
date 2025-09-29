-- +goose Up
ALTER TABLE hotels
    ADD COLUMN stars INT DEFAULT 0,            -- количество звёзд
    ADD COLUMN guests TEXT,                   -- кол-во человек в номере ("4-5 человек")
    ADD COLUMN distance_text TEXT,            -- человеко-читаемое расстояние
    ADD COLUMN photo_url TEXT;                -- ссылка на картинку

-- +goose Down
ALTER TABLE hotels
    DROP COLUMN IF EXISTS stars,
    DROP COLUMN IF EXISTS guests,
    DROP COLUMN IF EXISTS distance_text,
    DROP COLUMN IF EXISTS photo_url;
