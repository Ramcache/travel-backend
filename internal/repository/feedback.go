package repository

import (
	"context"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedbackRepo struct {
	db *pgxpool.Pool
}

func NewFeedbackRepo(db *pgxpool.Pool) *FeedbackRepo {
	return &FeedbackRepo{db: db}
}

func (r *FeedbackRepo) Create(ctx context.Context, f *models.Feedback) error {
	query := `INSERT INTO feedbacks (user_name, user_phone) 
              VALUES ($1, $2) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, f.UserName, f.UserPhone).Scan(&f.ID, &f.CreatedAt)
}

func (r *FeedbackRepo) List(ctx context.Context, limit, offset int) ([]models.Feedback, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, user_name, user_phone, is_read, created_at
        FROM feedbacks
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Feedback
	for rows.Next() {
		var f models.Feedback
		if err := rows.Scan(&f.ID, &f.UserName, &f.UserPhone, &f.IsRead, &f.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, nil
}

func (r *FeedbackRepo) MarkAsRead(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `UPDATE feedbacks SET is_read = true WHERE id = $1`, id)
	return err
}
