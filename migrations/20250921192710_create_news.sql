-- +goose Up
CREATE TABLE news (
                      id SERIAL PRIMARY KEY,
                      slug TEXT NOT NULL UNIQUE,
                      title TEXT NOT NULL,
                      excerpt TEXT,
                      content TEXT,
                      category TEXT NOT NULL CHECK (category IN ('hadj','company','other')),
                      media_type TEXT NOT NULL CHECK (media_type IN ('photo','video')),
                      preview_url TEXT,
                      video_url TEXT,
                      comments_count INT NOT NULL DEFAULT 0,
                      reposts_count INT NOT NULL DEFAULT 0,
                      views_count INT NOT NULL DEFAULT 0,
                      author_id INT REFERENCES users(id) ON DELETE SET NULL,
                      status TEXT NOT NULL DEFAULT 'published' CHECK (status IN ('draft','published','archived')),
                      published_at TIMESTAMP NOT NULL DEFAULT now(),
                      created_at TIMESTAMP NOT NULL DEFAULT now(),
                      updated_at TIMESTAMP NOT NULL DEFAULT now()
);


CREATE INDEX idx_news_category_published_at ON news(category, published_at DESC);
CREATE INDEX idx_news_status_published_at ON news(status, published_at DESC);


-- +goose Down
DROP TABLE news;