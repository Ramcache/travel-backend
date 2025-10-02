package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/Ramcache/travel-backend/internal/models"
)

type TripRepository struct {
	Db DB
}

func NewTripRepository(db DB) *TripRepository {
	return &TripRepository{Db: db}
}

type TripRepositoryI interface {
	List(ctx context.Context, f models.TripFilter) ([]models.Trip, error)
	GetByID(ctx context.Context, id int) (*models.Trip, error)
	Create(ctx context.Context, t *models.Trip) error
	Update(ctx context.Context, t *models.Trip) error
	Delete(ctx context.Context, id int) error
	GetMain(ctx context.Context) (*models.Trip, error)
	ResetMain(ctx context.Context, excludeID *int) error
	Popular(ctx context.Context, limit int) ([]models.Trip, error)
	IncrementViews(ctx context.Context, id int) error
	IncrementBuys(ctx context.Context, id int) error
	GetOptions(ctx context.Context, tripID int) ([]models.TripOptionResponse, error)
}

// List with filters (показываем только активные туры)
func (r *TripRepository) List(ctx context.Context, f models.TripFilter) ([]models.Trip, error) {
	filters := []string{"1=1"}
	args := []interface{}{}
	i := 1

	if f.Title != "" {
		filters = append(filters, fmt.Sprintf("title ILIKE $%d", i))
		args = append(args, "%"+f.Title+"%")
		i++
	}
	if f.DepartureCity != "" {
		filters = append(filters, fmt.Sprintf("departure_city = $%d", i))
		args = append(args, f.DepartureCity)
		i++
	}
	if f.TripType != "" {
		filters = append(filters, fmt.Sprintf("trip_type = $%d", i))
		args = append(args, f.TripType)
		i++
	}
	if f.Season != "" {
		filters = append(filters, fmt.Sprintf("season = $%d", i))
		args = append(args, f.Season)
		i++
	}
	if f.Active != nil {
		filters = append(filters, fmt.Sprintf("active = $%d", i))
		args = append(args, *f.Active)
		i++
	}
	if !f.StartAfter.IsZero() {
		filters = append(filters, fmt.Sprintf("start_date >= $%d", i))
		args = append(args, f.StartAfter)
		i++
	}
	if !f.EndBefore.IsZero() {
		filters = append(filters, fmt.Sprintf("end_date <= $%d", i))
		args = append(args, f.EndBefore)
		i++
	}
	if f.RouteCity != "" {
		filters = append(filters, fmt.Sprintf("EXISTS (SELECT 1 FROM trip_routes WHERE trip_id=trips.id AND city ILIKE $%d)", i))
		args = append(args, "%"+f.RouteCity+"%")
		i++
	}

	query := `SELECT id, title, description, photo_url, departure_city, trip_type, season,
                     price, discount_percent, currency,
                     start_date, end_date, booking_deadline, main, active,
                     views_count, buys_count, created_at, updated_at
              FROM trips WHERE ` + strings.Join(filters, " AND ") + `
              ORDER BY created_at DESC`

	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", i)
		args = append(args, f.Limit)
		i++
	}
	if f.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", i)
		args = append(args, f.Offset)
		i++
	}

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
			&t.DepartureCity, &t.TripType, &t.Season,
			&t.Price, &t.DiscountPercent, &t.Currency,
			&t.StartDate, &t.EndDate, &t.BookingDeadline,
			&t.Main, &t.Active,
			&t.ViewsCount, &t.BuysCount,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// расчет финальной цены
		if t.DiscountPercent > 0 {
			t.FinalPrice = t.Price * (100 - float64(t.DiscountPercent)) / 100
		} else {
			t.FinalPrice = t.Price
		}

		trips = append(trips, t)
	}
	return trips, rows.Err()
}

func (r *TripRepository) GetByID(ctx context.Context, id int) (*models.Trip, error) {
	var t models.Trip
	err := r.Db.QueryRow(ctx,
		`SELECT id, title, description, photo_url, departure_city, trip_type, season,
                price, discount_percent, currency,
                start_date, end_date, booking_deadline, main, active,
                views_count, buys_count, created_at, updated_at
         FROM trips WHERE id=$1`, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.PhotoURL,
			&t.DepartureCity, &t.TripType, &t.Season,
			&t.Price, &t.DiscountPercent, &t.Currency,
			&t.StartDate, &t.EndDate, &t.BookingDeadline,
			&t.Main, &t.Active,
			&t.ViewsCount, &t.BuysCount,
			&t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// расчет финальной цены
	if t.DiscountPercent > 0 {
		t.FinalPrice = t.Price * (100 - float64(t.DiscountPercent)) / 100
	} else {
		t.FinalPrice = t.Price
	}

	// подтягиваем отели
	rows, err := r.Db.Query(ctx, `
        SELECT h.id, h.name, h.city, h.distance, h.meals, h.rating, th.nights
        FROM trip_hotels th
        JOIN hotels h ON h.id = th.hotel_id
        WHERE th.trip_id = $1
        ORDER BY h.city`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var h models.TripHotelWithInfo
		if err := rows.Scan(&h.HotelID, &h.Name, &h.City, &h.Distance, &h.Meals, &h.Rating, &h.Nights); err != nil {
			return nil, err
		}
		t.Hotels = append(t.Hotels, h)
	}

	return &t, nil
}

func (r *TripRepository) Create(ctx context.Context, t *models.Trip) error {
	return r.Db.QueryRow(ctx,
		`INSERT INTO trips (title, description, photo_url, departure_city, trip_type, season,
                            price, discount_percent, currency,
                            start_date, end_date, booking_deadline, main, active)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
         RETURNING id, views_count, buys_count, created_at, updated_at`,
		t.Title, t.Description, t.PhotoURL, t.DepartureCity, t.TripType, t.Season,
		t.Price, t.DiscountPercent, t.Currency,
		t.StartDate, t.EndDate, t.BookingDeadline, t.Main, t.Active,
	).Scan(&t.ID, &t.ViewsCount, &t.BuysCount, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TripRepository) Update(ctx context.Context, t *models.Trip) error {
	err := r.Db.QueryRow(ctx,
		`UPDATE trips
         SET title=$1, description=$2, photo_url=$3, departure_city=$4, trip_type=$5, season=$6,
             price=$7, discount_percent=$8, currency=$9,
             start_date=$10, end_date=$11, booking_deadline=$12, main=$13, active=$14, updated_at=now()
         WHERE id=$15
         RETURNING views_count, buys_count, updated_at`,
		t.Title, t.Description, t.PhotoURL, t.DepartureCity, t.TripType, t.Season,
		t.Price, t.DiscountPercent, t.Currency,
		t.StartDate, t.EndDate, t.BookingDeadline,
		t.Main, t.Active, t.ID,
	).Scan(&t.ViewsCount, &t.BuysCount, &t.UpdatedAt)
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
		`SELECT id, title, description, photo_url, departure_city, trip_type, season,
                price, discount_percent, currency,
                start_date, end_date, booking_deadline, main, active,
                views_count, buys_count, created_at, updated_at
         FROM trips WHERE main = true AND active = true LIMIT 1`).
		Scan(&t.ID, &t.Title, &t.Description, &t.PhotoURL,
			&t.DepartureCity, &t.TripType, &t.Season,
			&t.Price, &t.DiscountPercent, &t.Currency,
			&t.StartDate, &t.EndDate, &t.BookingDeadline,
			&t.Main, &t.Active,
			&t.ViewsCount, &t.BuysCount,
			&t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// расчет финальной цены
	if t.DiscountPercent > 0 {
		t.FinalPrice = t.Price * (100 - float64(t.DiscountPercent)) / 100
	} else {
		t.FinalPrice = t.Price
	}

	return &t, nil
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

// Popular — топ туров по количеству покупок (только активные)
func (r *TripRepository) Popular(ctx context.Context, limit int) ([]models.Trip, error) {
	rows, err := r.Db.Query(ctx, `
        SELECT id, title, description, photo_url, departure_city, trip_type, season,
               price, discount_percent, currency,
               start_date, end_date, booking_deadline, main, active,
               views_count, buys_count, created_at, updated_at
        FROM trips
        WHERE active = true
        ORDER BY buys_count DESC, views_count DESC
        LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []models.Trip
	for rows.Next() {
		var t models.Trip
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.PhotoURL,
			&t.DepartureCity, &t.TripType, &t.Season,
			&t.Price, &t.DiscountPercent, &t.Currency,
			&t.StartDate, &t.EndDate, &t.BookingDeadline,
			&t.Main, &t.Active,
			&t.ViewsCount, &t.BuysCount,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// расчет финальной цены
		if t.DiscountPercent > 0 {
			t.FinalPrice = t.Price * (100 - float64(t.DiscountPercent)) / 100
		} else {
			t.FinalPrice = t.Price
		}

		trips = append(trips, t)
	}
	return trips, rows.Err()
}

func (r *TripRepository) IncrementViews(ctx context.Context, id int) error {
	_, err := r.Db.Exec(ctx,
		`UPDATE trips SET views_count = views_count + 1 WHERE id = $1`,
		id,
	)
	return err
}

func (r *TripRepository) IncrementBuys(ctx context.Context, id int) error {
	_, err := r.Db.Exec(ctx,
		`UPDATE trips SET buys_count = buys_count + 1 WHERE id = $1`,
		id,
	)
	return err
}

func (r *TripRepository) GetOptions(ctx context.Context, tripID int) ([]models.TripOptionResponse, error) {
	rows, err := r.Db.Query(ctx,
		`SELECT id, name, price, unit FROM trip_options WHERE trip_id=$1 ORDER BY id`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var opts []models.TripOptionResponse
	for rows.Next() {
		var o models.TripOptionResponse
		if err := rows.Scan(&o.ID, &o.Name, &o.Price, &o.Unit); err != nil {
			return nil, err
		}
		opts = append(opts, o)
	}
	return opts, rows.Err()
}
