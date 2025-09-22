-- +goose Up
CREATE INDEX IF NOT EXISTS idx_news_views_count_published_at
    ON news (views_count DESC, published_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_news_views_count_published_at;
