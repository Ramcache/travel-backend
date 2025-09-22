-- +goose Up
CREATE OR REPLACE VIEW v_news_by_category AS
SELECT COALESCE(nc.title, 'Без категории') AS category, COUNT(n.id) AS cnt
FROM news n
         LEFT JOIN news_categories nc ON nc.id = n.category_id
GROUP BY COALESCE(nc.title, 'Без категории');

-- +goose Down
CREATE OR REPLACE VIEW v_news_by_category AS
SELECT nc.title AS category, COUNT(n.id) AS cnt
FROM news n
         LEFT JOIN news_categories nc ON nc.id = n.category_id
GROUP BY nc.title;
