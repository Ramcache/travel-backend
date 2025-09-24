-- +goose Up
CREATE TABLE orders (
                        id SERIAL PRIMARY KEY,
                        trip_id INT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
                        user_name TEXT NOT NULL,
                        user_phone TEXT NOT NULL,
                        status TEXT DEFAULT 'new',
                        created_at TIMESTAMP DEFAULT now()
);

-- +goose Down
DROP TABLE orders;
