-- +goose Up
CREATE TABLE trip_routes (
                             id SERIAL PRIMARY KEY,
                             trip_id INT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
                             city VARCHAR(255) NOT NULL,
                             transport VARCHAR(50),
                             duration VARCHAR(50),
                             position INT NOT NULL,
                             created_at TIMESTAMP NOT NULL DEFAULT now(),
                             updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS trip_routes;
