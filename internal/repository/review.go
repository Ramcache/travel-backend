package repository

import (
	"context"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepo struct {
	db *pgxpool.Pool
}

func NewReviewRepo(db *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{db: db}
}

func (r *ReviewRepo) Create(ctx context.Context, rev *models.TripReview) error {
	return r.db.QueryRow(ctx, `
        INSERT INTO trip_reviews (trip_id, user_name, rating, comment)
        VALUES ($1,$2,$3,$4)
        RETURNING id, created_at`,
		rev.TripID, rev.UserName, rev.Rating, rev.Comment,
	).Scan(&rev.ID, &rev.CreatedAt)
}

func (r *ReviewRepo) ListByTrip(ctx context.Context, tripID, limit, offset int) ([]models.TripReview, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM trip_reviews WHERE trip_id=$1`, tripID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
        SELECT id, trip_id, user_name, rating, comment, created_at
        FROM trip_reviews
        WHERE trip_id=$1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3`, tripID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []models.TripReview
	for rows.Next() {
		var r models.TripReview
		if err := rows.Scan(&r.ID, &r.TripID, &r.UserName, &r.Rating, &r.Comment, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, r)
	}
	return reviews, total, rows.Err()
}
