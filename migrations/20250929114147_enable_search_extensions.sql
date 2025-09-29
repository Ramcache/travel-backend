-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- индексы для быстрого поиска (без unaccent в выражении)
CREATE INDEX IF NOT EXISTS idx_trips_search
    ON trips USING gin (to_tsvector('simple', title || ' ' || coalesce(description, '')));

CREATE INDEX IF NOT EXISTS idx_news_search
    ON news USING gin (to_tsvector('simple', title || ' ' || coalesce(content, '')));

-- +goose Down
DROP INDEX IF EXISTS idx_trips_search;
DROP INDEX IF EXISTS idx_news_search;

DROP EXTENSION IF EXISTS unaccent;
DROP EXTENSION IF EXISTS pg_trgm;
