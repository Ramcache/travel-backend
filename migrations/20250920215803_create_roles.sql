-- +goose Up
CREATE TABLE roles (
                       id SERIAL PRIMARY KEY,
                       name TEXT UNIQUE NOT NULL
);

INSERT INTO roles (name) VALUES
                             ('user'),
                             ('admin'),
                             ('manager');

-- +goose Down
DROP TABLE roles;
