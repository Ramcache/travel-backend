-- +goose Up
CREATE TABLE hotels (
                        id SERIAL PRIMARY KEY,
                        name TEXT NOT NULL,
                        city TEXT NOT NULL,
                        distance NUMERIC,
                        meals TEXT,
                        rating INT DEFAULT 0,
                        created_at TIMESTAMP DEFAULT now(),
                        updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE trip_hotels (
                             trip_id INT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
                             hotel_id INT NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
                             nights INT NOT NULL DEFAULT 0,
                             PRIMARY KEY (trip_id, hotel_id)
);

-- +goose Down
DROP TABLE IF EXISTS trip_hotels;
DROP TABLE IF EXISTS hotels;
