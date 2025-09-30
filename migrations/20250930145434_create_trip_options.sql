-- +goose Up
CREATE TABLE trip_options (
                              id SERIAL PRIMARY KEY,
                              trip_id INT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
                              name TEXT NOT NULL,
                              price NUMERIC(12,2) NOT NULL,
                              unit TEXT DEFAULT 'per_day',
                              created_at TIMESTAMP DEFAULT now()
);

-- +goose Down
DROP TABLE trip_options;
