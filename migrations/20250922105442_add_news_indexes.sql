-- +goose Up
-- Индекс для последних опубликованных новостей (recent, list)
CREATE INDEX IF NOT EXISTS idx_news_status_published_at_id_desc
    ON news (status, published_at DESC, id DESC);

-- Индекс для поиска по slug (деталка новости)
CREATE UNIQUE INDEX IF NOT EXISTS idx_news_slug_lower
    ON news (LOWER(slug));

-- Индекс для фильтрации по category/media_type
CREATE INDEX IF NOT EXISTS idx_news_category_media_type
    ON news (category, media_type);

-- +goose Down
DROP INDEX IF EXISTS idx_news_status_published_at_id_desc;
DROP INDEX IF EXISTS idx_news_slug_lower;
DROP INDEX IF EXISTS idx_news_category_media_type;
