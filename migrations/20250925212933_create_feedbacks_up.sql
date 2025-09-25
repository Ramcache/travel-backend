-- +goose Up
CREATE TABLE feedbacks (
                           id SERIAL PRIMARY KEY,
                           user_name  TEXT NOT NULL,
                           user_phone TEXT NOT NULL,
                           is_read    BOOLEAN DEFAULT false,
                           created_at TIMESTAMP DEFAULT now()
);

-- +goose Down
DROP TABLE feedbacks;
