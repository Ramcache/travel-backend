-- +goose Up
CREATE OR REPLACE VIEW v_news_by_status AS
SELECT status, COUNT(*) AS cnt
FROM news
GROUP BY status;

CREATE OR REPLACE VIEW v_news_by_category AS
SELECT nc.title AS category, COUNT(n.id) AS cnt
FROM news n
         LEFT JOIN news_categories nc ON nc.id = n.category_id
GROUP BY nc.title;

CREATE OR REPLACE VIEW v_trips_by_type AS
SELECT trip_type, COUNT(*) AS cnt
FROM trips
GROUP BY trip_type;

CREATE OR REPLACE VIEW v_trips_by_city AS
SELECT departure_city, COUNT(*) AS cnt
FROM trips
GROUP BY departure_city;

CREATE OR REPLACE VIEW v_users_by_role AS
SELECT r.name AS role, COUNT(u.id) AS cnt
FROM roles r
         LEFT JOIN users u ON u.role_id = r.id
GROUP BY r.name;

-- +goose Down
DROP VIEW IF EXISTS v_users_by_role;
DROP VIEW IF EXISTS v_trips_by_city;
DROP VIEW IF EXISTS v_trips_by_type;
DROP VIEW IF EXISTS v_news_by_category;
DROP VIEW IF EXISTS v_news_by_status;
