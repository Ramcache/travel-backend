package repository

import (
	"context"

	"github.com/Ramcache/travel-backend/internal/models"
)

type TripRouteRepository interface {
	Create(ctx context.Context, tripID int, req models.TripRouteRequest) (*models.TripRoute, error)
	ListByTrip(ctx context.Context, tripID int) ([]models.TripRoute, error)
	Update(ctx context.Context, id int, req models.TripRouteRequest) (*models.TripRoute, error)
	Delete(ctx context.Context, id int) error
}

type tripRouteRepo struct {
	pool DB
}

func NewTripRouteRepository(pool DB) TripRouteRepository {
	return &tripRouteRepo{pool: pool}
}

func (r *tripRouteRepo) ListByTrip(ctx context.Context, tripID int) ([]models.TripRoute, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, trip_id, city, transport, duration, position, created_at, updated_at
		FROM trip_routes
		WHERE trip_id = $1
		ORDER BY position ASC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.TripRoute
	for rows.Next() {
		var rt models.TripRoute
		if err := rows.Scan(&rt.ID, &rt.TripID, &rt.City, &rt.Transport, &rt.Duration, &rt.Position, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, rt)
	}
	return routes, nil
}

func (r *tripRouteRepo) Create(ctx context.Context, tripID int, req models.TripRouteRequest) (*models.TripRoute, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO trip_routes (trip_id, city, transport, duration, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, trip_id, city, transport, duration, position, created_at, updated_at
	`, tripID, req.City, req.Transport, req.Duration, req.Position)

	var rt models.TripRoute
	if err := row.Scan(&rt.ID, &rt.TripID, &rt.City, &rt.Transport, &rt.Duration, &rt.Position, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *tripRouteRepo) Update(ctx context.Context, id int, req models.TripRouteRequest) (*models.TripRoute, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE trip_routes
		SET city = $1, transport = $2, duration = $3, position = $4, updated_at = now()
		WHERE id = $5
		RETURNING id, trip_id, city, transport, duration, position, created_at, updated_at
	`, req.City, req.Transport, req.Duration, req.Position, id)

	var rt models.TripRoute
	if err := row.Scan(&rt.ID, &rt.TripID, &rt.City, &rt.Transport, &rt.Duration, &rt.Position, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *tripRouteRepo) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM trip_routes WHERE id = $1`, id)
	return err
}
