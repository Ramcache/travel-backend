package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ramcache/travel-backend/internal/models"
)

type TripRepository struct {
	Db *pgxpool.Pool
}

func NewTripRepository(db *pgxpool.Pool) *TripRepository {
	return &TripRepository{Db: db}
}

// List with filters
func (r *TripRepository) List(ctx context.Context, departureCity, tripType, season string) ([]models.Trip, error) {
	filters := []string{}
	args := []interface{}{}
	i := 1

	if departureCity != "" {
		filters = append(filters, fmt.Sprintf("departure_city = $%d", i))
		args = append(args, departureCity)
		i++
	}
	if tripType != "" {
		filters = append(filters, fmt.Sprintf("trip_type = $%d", i))
		args = append(args, tripType)
		i++
	}
	if season != "" {
		filters = append(filters, fmt.Sprintf("season = $%d", i))
		args = append(args, season)
		i++
	}

	query := `SELECT id, title, description, photo_url, departure_city, trip_type, season, price, currency,
                     start_date, end_date, booking_deadline, main, created_at, updated_at FROM trips`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.Db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []models.Trip
	for rows.Next() {
		var t models.Trip
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.PhotoURL,
			&t.DepartureCity, &t.TripType, &t.Season, &t.Price, &t.Currency,
			&t.StartDate, &t.EndDate, &t.BookingDeadline, &t.Main, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		trips = append(trips, t)
	}
	return trips, rows.Err()
}

func (r *TripRepository) GetByID(ctx context.Context, id int) (*models.Trip, error) {
	var t models.Trip
	err := r.Db.QueryRow(ctx,
		`SELECT id, title, description, photo_url, departure_city, trip_type, season, price, currency,
                start_date, end_date, booking_deadline, main, created_at, updated_at
         FROM trips WHERE id=$1`, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.PhotoURL, &t.DepartureCity, &t.TripType, &t.Season,
			&t.Price, &t.Currency, &t.StartDate, &t.EndDate, &t.BookingDeadline, &t.Main, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TripRepository) Create(ctx context.Context, t *models.Trip) error {
	return r.Db.QueryRow(ctx,
		`INSERT INTO trips (title, description, photo_url, departure_city, trip_type, season, price, currency,
                            start_date, end_date, booking_deadline, main)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
         RETURNING id, created_at, updated_at`,
		t.Title, t.Description, t.PhotoURL, t.DepartureCity, t.TripType, t.Season,
		t.Price, t.Currency, t.StartDate, t.EndDate, t.BookingDeadline, t.Main,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TripRepository) Update(ctx context.Context, t *models.Trip) error {
	err := r.Db.QueryRow(ctx,
		`UPDATE trips
         SET title=$1, description=$2, photo_url=$3, departure_city=$4, trip_type=$5, season=$6,
             price=$7, currency=$8, start_date=$9, end_date=$10, booking_deadline=$11, main=$12, updated_at=now()
         WHERE id=$13
         RETURNING updated_at`,
		t.Title, t.Description, t.PhotoURL, t.DepartureCity, t.TripType, t.Season,
		t.Price, t.Currency, t.StartDate, t.EndDate, t.BookingDeadline, t.Main, t.ID,
	).Scan(&t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	return err
}

func (r *TripRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.Db.Exec(ctx, `DELETE FROM trips WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TripRepository) GetMain(ctx context.Context) (*models.Trip, error) {
	var t models.Trip
	err := r.Db.QueryRow(ctx,
		`SELECT id, title, description, photo_url, departure_city, trip_type, season, price, currency,
                start_date, end_date, booking_deadline, main, created_at, updated_at
         FROM trips WHERE main = true LIMIT 1`).
		Scan(&t.ID, &t.Title, &t.Description, &t.PhotoURL, &t.DepartureCity, &t.TripType, &t.Season,
			&t.Price, &t.Currency, &t.StartDate, &t.EndDate, &t.BookingDeadline, &t.Main, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return &t, err
}

// ResetMain сбрасывает main у всех туров
func (r *TripRepository) ResetMain(ctx context.Context, excludeID *int) error {
	if excludeID != nil {
		_, err := r.Db.Exec(ctx, `UPDATE trips SET main=false WHERE id <> $1`, *excludeID)
		return err
	}
	_, err := r.Db.Exec(ctx, `UPDATE trips SET main=false`)
	return err
}
