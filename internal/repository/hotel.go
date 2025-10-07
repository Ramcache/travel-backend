package repository

import (
	"context"

	"github.com/Ramcache/travel-backend/internal/models"
)

// ======== Интерфейс ========
type HotelRepositoryI interface {
	Create(ctx context.Context, hotel *models.Hotel) error
	Get(ctx context.Context, id int) (*models.Hotel, error)
	List(ctx context.Context) ([]models.Hotel, error)
	Update(ctx context.Context, hotel *models.Hotel) error
	Delete(ctx context.Context, id int) error
	Attach(ctx context.Context, th *models.TripHotel) error
	ListByTrip(ctx context.Context, tripID int) ([]models.Hotel, error)
	ClearByTrip(ctx context.Context, tripID int) (int64, error)
}

// ======== Реализация ========
type HotelRepository struct {
	db DB
}

func NewHotelRepository(db DB) *HotelRepository {
	return &HotelRepository{db: db}
}

// список полей для SELECT
const hotelFields = `
	id, name, city, stars, distance, distance_text, meals, guests, urls, transfer, created_at, updated_at
`

func scanHotel(row interface{ Scan(dest ...any) error }) (models.Hotel, error) {
	var h models.Hotel
	err := row.Scan(
		&h.ID,
		&h.Name,
		&h.City,
		&h.Stars,
		&h.Distance,
		&h.DistanceText,
		&h.Meals,
		&h.Guests,
		&h.URLs, // 👈 массив ссылок (TEXT[])
		&h.Transfer,
		&h.CreatedAt,
		&h.UpdatedAt,
	)
	return h, err
}

// ======== CRUD ========

// Create hotel
func (r *HotelRepository) Create(ctx context.Context, hotel *models.Hotel) error {
	query := `
		INSERT INTO hotels (name, city, stars, distance, distance_text, meals, guests, urls, transfer)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		hotel.Name,
		hotel.City,
		hotel.Stars,
		hotel.Distance,
		hotel.DistanceText,
		hotel.Meals,
		hotel.Guests,
		hotel.URLs,
		hotel.Transfer,
	).Scan(&hotel.ID, &hotel.CreatedAt, &hotel.UpdatedAt)
}

// Get hotel by ID
func (r *HotelRepository) Get(ctx context.Context, id int) (*models.Hotel, error) {
	query := `SELECT ` + hotelFields + ` FROM hotels WHERE id=$1`
	h, err := scanHotel(r.db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &h, nil
}

// List all hotels
func (r *HotelRepository) List(ctx context.Context) ([]models.Hotel, error) {
	query := `SELECT ` + hotelFields + ` FROM hotels ORDER BY id DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotels []models.Hotel
	for rows.Next() {
		h, err := scanHotel(rows)
		if err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}
	return hotels, rows.Err()
}

// Update hotel
func (r *HotelRepository) Update(ctx context.Context, hotel *models.Hotel) error {
	query := `
		UPDATE hotels 
		SET name=$1, city=$2, stars=$3, distance=$4, distance_text=$5,
		    meals=$6, guests=$7, urls=$8, transfer=$9, updated_at=now()
		WHERE id=$10
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, query,
		hotel.Name,
		hotel.City,
		hotel.Stars,
		hotel.Distance,
		hotel.DistanceText,
		hotel.Meals,
		hotel.Guests,
		hotel.URLs, // 👈 TEXT[]
		hotel.Transfer,
		hotel.ID,
	).Scan(&hotel.UpdatedAt)
	if err != nil {
		return mapNotFound(err)
	}
	return nil
}

// Delete hotel
func (r *HotelRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM hotels WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ======== Trip ↔ Hotel ========

// Attach hotel to trip
func (r *HotelRepository) Attach(ctx context.Context, th *models.TripHotel) error {
	query := `
		INSERT INTO trip_hotels (trip_id, hotel_id, nights)
		VALUES ($1,$2,$3)
		ON CONFLICT (trip_id, hotel_id) DO UPDATE 
		SET nights = EXCLUDED.nights
	`
	_, err := r.db.Exec(ctx, query, th.TripID, th.HotelID, th.Nights)
	return err
}

// List hotels by trip
func (r *HotelRepository) ListByTrip(ctx context.Context, tripID int) ([]models.Hotel, error) {
	query := `
		SELECT h.` + hotelFields + `, th.nights
		FROM trip_hotels th
		JOIN hotels h ON h.id = th.hotel_id
		WHERE th.trip_id = $1
		ORDER BY h.city
	`
	rows, err := r.db.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotels []models.Hotel
	for rows.Next() {
		var h models.Hotel
		if err := rows.Scan(
			&h.ID,
			&h.Name,
			&h.City,
			&h.Stars,
			&h.Distance,
			&h.DistanceText,
			&h.Meals,
			&h.Guests,
			&h.URLs, // 👈 TEXT[]
			&h.Transfer,
			&h.CreatedAt,
			&h.UpdatedAt,
			&h.Nights,
		); err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}
	return hotels, rows.Err()
}

// Clear hotels from trip
func (r *HotelRepository) ClearByTrip(ctx context.Context, tripID int) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM trip_hotels WHERE trip_id=$1`, tripID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
