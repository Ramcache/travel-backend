-- +goose Up
CREATE TABLE news_categories (
                                 id SERIAL PRIMARY KEY,
                                 slug TEXT UNIQUE NOT NULL,
                                 title TEXT NOT NULL,
                                 created_at TIMESTAMP DEFAULT now(),
                                 updated_at TIMESTAMP DEFAULT now()
);

ALTER TABLE news
    ADD COLUMN category_id INT,
    ADD CONSTRAINT fk_news_category FOREIGN KEY (category_id) REFERENCES news_categories(id);

-- перенос старых данных из text-колонки category (если она была)
-- UPDATE news SET category_id = (SELECT id FROM news_categories WHERE slug = news.category);

-- ALTER TABLE news DROP COLUMN category;


-- +goose Down
-- вернуть как было
ALTER TABLE news DROP CONSTRAINT IF EXISTS fk_news_category;
ALTER TABLE news DROP COLUMN IF EXISTS category_id;

DROP TABLE IF EXISTS news_categories;
