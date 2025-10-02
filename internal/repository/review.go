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

// общий SELECT список
const reviewFields = `
	id, trip_id, user_name, rating, comment, created_at
`

// сканер отзыва
func scanReview(row interface{ Scan(dest ...any) error }) (models.TripReview, error) {
	var r models.TripReview
	err := row.Scan(&r.ID, &r.TripID, &r.UserName, &r.Rating, &r.Comment, &r.CreatedAt)
	return r, err
}

func (r *ReviewRepo) Create(ctx context.Context, rev *models.TripReview) error {
	query := `INSERT INTO trip_reviews (trip_id, user_name, rating, comment)
              VALUES ($1,$2,$3,$4)
              RETURNING id, created_at`
	return r.db.QueryRow(ctx, query,
		rev.TripID, rev.UserName, rev.Rating, rev.Comment,
	).Scan(&rev.ID, &rev.CreatedAt)
}

func (r *ReviewRepo) ListByTrip(ctx context.Context, tripID, limit, offset int) ([]models.TripReview, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM trip_reviews WHERE trip_id=$1`, tripID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT ` + reviewFields + `
              FROM trip_reviews
              WHERE trip_id=$1
              ORDER BY created_at DESC
              LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, tripID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []models.TripReview
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, total, rows.Err()
}
