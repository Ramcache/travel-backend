-- +goose Up
CREATE TABLE trip_reviews (
                              id SERIAL PRIMARY KEY,
                              trip_id INT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
                              user_name TEXT NOT NULL,
                              rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
                              comment TEXT,
                              created_at TIMESTAMP DEFAULT now()
);

-- +goose Down
DROP TABLE trip_reviews;
