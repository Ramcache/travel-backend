package repository

import (
	"context"
	"github.com/Ramcache/travel-backend/internal/models"
	"strconv"
)

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

type HotelRepository struct {
	db DB // твой pgx pool wrapper
}

func NewHotelRepository(db DB) *HotelRepository {
	return &HotelRepository{db: db}
}

func (r *HotelRepository) Create(ctx context.Context, hotel *models.Hotel) error {
	query := `INSERT INTO hotels (name, city, stars, distance, distance_text, meals, guests, photo_url) 
              VALUES ($1,$2,$3,$4,$5,$6,$7,$8) 
              RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query,
		hotel.Name, hotel.City, hotel.Stars, hotel.Distance, hotel.DistanceText,
		hotel.Meals, hotel.Guests, hotel.PhotoURL).
		Scan(&hotel.ID, &hotel.CreatedAt, &hotel.UpdatedAt)
}

func (r *HotelRepository) Get(ctx context.Context, id int) (*models.Hotel, error) {
	var h models.Hotel
	query := `SELECT id, name, city, stars, distance, distance_text, meals, guests, photo_url, created_at, updated_at 
              FROM hotels WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&h.ID, &h.Name, &h.City, &h.Stars, &h.Distance, &h.DistanceText,
		&h.Meals, &h.Guests, &h.PhotoURL, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HotelRepository) List(ctx context.Context) ([]models.Hotel, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, city, stars, distance, distance_text, meals, guests, photo_url, created_at, updated_at FROM hotels ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotels []models.Hotel
	for rows.Next() {
		var h models.Hotel
		if err := rows.Scan(&h.ID, &h.Name, &h.City, &h.Stars, &h.Distance, &h.DistanceText,
			&h.Meals, &h.Guests, &h.PhotoURL, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}
	return hotels, nil
}

func (r *HotelRepository) Update(ctx context.Context, hotel *models.Hotel) error {
	query := `UPDATE hotels 
              SET name=$1, city=$2, stars=$3, distance=$4, distance_text=$5, meals=$6, guests=$7, photo_url=$8, updated_at=now() 
              WHERE id=$9`
	_, err := r.db.Exec(ctx, query,
		hotel.Name, hotel.City, hotel.Stars, hotel.Distance, hotel.DistanceText,
		hotel.Meals, hotel.Guests, hotel.PhotoURL, hotel.ID)
	return err
}

func (r *HotelRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM hotels WHERE id=$1`, id)
	return err
}

func (r *HotelRepository) Attach(ctx context.Context, th *models.TripHotel) error {
	query := `INSERT INTO trip_hotels (trip_id, hotel_id, nights) VALUES ($1,$2,$3)
              ON CONFLICT (trip_id, hotel_id) DO UPDATE SET nights = EXCLUDED.nights`
	_, err := r.db.Exec(ctx, query, th.TripID, th.HotelID, th.Nights)
	return err
}

func (r *HotelRepository) ListByTrip(ctx context.Context, tripID int) ([]models.Hotel, error) {
	rows, err := r.db.Query(ctx, `
        SELECT h.id, h.name, h.city, h.stars, h.distance, h.distance_text, h.meals, h.guests, h.photo_url, h.created_at, h.updated_at, th.nights
        FROM trip_hotels th
        JOIN hotels h ON h.id = th.hotel_id
        WHERE th.trip_id = $1
    `, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hotels []models.Hotel
	for rows.Next() {
		var h models.Hotel
		var nights int
		if err := rows.Scan(&h.ID, &h.Name, &h.City, &h.Stars, &h.Distance, &h.DistanceText,
			&h.Meals, &h.Guests, &h.PhotoURL, &h.CreatedAt, &h.UpdatedAt, &nights); err != nil {
			return nil, err
		}
		// пока ночи просто добавляем к структуре (можно расширить Hotel, добавив Nights)
		h.Meals = h.Meals + " (" + strconv.Itoa(nights) + " ночей)"
		hotels = append(hotels, h)
	}
	return hotels, nil
}

func (r *HotelRepository) ClearByTrip(ctx context.Context, tripID int) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM trip_hotels WHERE trip_id=$1`, tripID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
