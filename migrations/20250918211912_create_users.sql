-- +goose Up
CREATE TABLE users (
                       id          SERIAL PRIMARY KEY,
                       email       TEXT UNIQUE NOT NULL,
                       password    TEXT NOT NULL,
                       full_name   TEXT,
                       created_at  TIMESTAMPTZ DEFAULT now(),
                       updated_at  TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE users;
