package repository

import (
	"context"

	"github.com/Ramcache/travel-backend/internal/models"
)

type TripRouteRepository interface {
	Create(ctx context.Context, r *models.TripRoute) error
	ListByTrip(ctx context.Context, tripID int) ([]models.TripRoute, error)
	Update(ctx context.Context, id int, req models.TripRouteRequest) (*models.TripRoute, error)
	Delete(ctx context.Context, id int) error
	ClearByTrip(ctx context.Context, tripID int) (int64, error)
}

type tripRouteRepo struct {
	pool DB
}

func NewTripRouteRepository(pool DB) TripRouteRepository {
	return &tripRouteRepo{pool: pool}
}

const tripRouteFields = `
	id, trip_id, city, transport, duration, stop_time, position, created_at, updated_at
`

func scanTripRoute(row interface{ Scan(dest ...any) error }) (models.TripRoute, error) {
	var rt models.TripRoute
	err := row.Scan(&rt.ID, &rt.TripID, &rt.City, &rt.Transport, &rt.Duration,
		&rt.StopTime, &rt.Position, &rt.CreatedAt, &rt.UpdatedAt)
	return rt, err
}

func (r *tripRouteRepo) ListByTrip(ctx context.Context, tripID int) ([]models.TripRoute, error) {
	query := `SELECT ` + tripRouteFields + ` FROM trip_routes WHERE trip_id = $1 ORDER BY position ASC`
	rows, err := r.pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.TripRoute
	for rows.Next() {
		rt, err := scanTripRoute(rows)
		if err != nil {
			return nil, err
		}
		routes = append(routes, rt)
	}
	return routes, rows.Err()
}

// Create — создание маршрута
func (r *tripRouteRepo) Create(ctx context.Context, rt *models.TripRoute) error {
	if rt.Position == 0 {
		err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(position), 0) + 1 FROM trip_routes WHERE trip_id = $1`, rt.TripID).
			Scan(&rt.Position)
		if err != nil {
			return err
		}
	}

	query := `INSERT INTO trip_routes (trip_id, city, transport, duration, stop_time, position)
	          VALUES ($1,$2,$3,$4,$5,$6)
	          RETURNING ` + tripRouteFields

	row := r.pool.QueryRow(ctx, query, rt.TripID, rt.City, rt.Transport, rt.Duration, rt.StopTime, rt.Position)
	newRt, err := scanTripRoute(row)
	if err != nil {
		return err
	}
	*rt = newRt
	return nil
}

func (r *tripRouteRepo) Update(ctx context.Context, id int, req models.TripRouteRequest) (*models.TripRoute, error) {
	query := `UPDATE trip_routes
	          SET city=$1, transport=$2, duration=$3, stop_time=$4, position=$5, updated_at=now()
	          WHERE id=$6
	          RETURNING ` + tripRouteFields

	row := r.pool.QueryRow(ctx, query, req.City, req.Transport, req.Duration, req.StopTime, req.Position, id)
	rt, err := scanTripRoute(row)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &rt, nil
}

func (r *tripRouteRepo) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM trip_routes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ClearByTrip — удаляет все маршруты, связанные с конкретным туром
func (r *tripRouteRepo) ClearByTrip(ctx context.Context, tripID int) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM trip_routes WHERE trip_id = $1`, tripID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
