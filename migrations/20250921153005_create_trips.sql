-- +goose Up
CREATE TABLE trips (
                       id SERIAL PRIMARY KEY,
                       title TEXT NOT NULL,
                       description TEXT,
                       photo_url TEXT,
                       departure_city TEXT NOT NULL,
                       trip_type TEXT NOT NULL,
                       season TEXT,
                       price NUMERIC(12,2) NOT NULL,
                       currency TEXT NOT NULL DEFAULT 'RUB',
                       start_date DATE NOT NULL,
                       end_date DATE NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT now(),
                       updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE trips;
