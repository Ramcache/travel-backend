package repository

import (
	"context"
	"fmt"
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

func (r *FeedbackRepo) Count(ctx context.Context, phone string, isRead *bool) (int, error) {
	query := `SELECT COUNT(*) FROM feedbacks WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if phone != "" {
		query += fmt.Sprintf(" AND user_phone ILIKE $%d", argID)
		args = append(args, "%"+phone+"%")
		argID++
	}
	if isRead != nil {
		query += fmt.Sprintf(" AND is_read = $%d", argID)
		args = append(args, *isRead)
		argID++
	}

	var total int
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *FeedbackRepo) List(ctx context.Context, limit, offset int, phone string, isRead *bool) ([]models.Feedback, error) {
	query := `
        SELECT id, user_name, user_phone, is_read, created_at
        FROM feedbacks
        WHERE 1=1
    `
	args := []interface{}{}
	argID := 1

	if phone != "" {
		query += fmt.Sprintf(" AND user_phone ILIKE $%d", argID)
		args = append(args, "%"+phone+"%")
		argID++
	}
	if isRead != nil {
		query += fmt.Sprintf(" AND is_read = $%d", argID)
		args = append(args, *isRead)
		argID++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
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
