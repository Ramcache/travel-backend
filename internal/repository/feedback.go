package repository

import (
	"context"
	"fmt"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedbackRepo struct {
	db *pgxpool.Pool
}

func NewFeedbackRepo(db *pgxpool.Pool) *FeedbackRepo {
	return &FeedbackRepo{db: db}
}

// общий SELECT список
const feedbackFields = `
	id, user_name, user_phone, is_read, created_at
`

// приватный сканер
func scanFeedback(row pgx.Row) (models.Feedback, error) {
	var f models.Feedback
	err := row.Scan(&f.ID, &f.UserName, &f.UserPhone, &f.IsRead, &f.CreatedAt)
	return f, err
}

// buildFeedbackFilters собирает WHERE + args
func buildFeedbackFilters(phone string, isRead *bool) (string, []any) {
	filters := "1=1"
	args := []any{}
	i := 1

	if phone != "" {
		filters += fmt.Sprintf(" AND user_phone ILIKE $%d", i)
		args = append(args, "%"+phone+"%")
		i++
	}
	if isRead != nil {
		filters += fmt.Sprintf(" AND is_read = $%d", i)
		args = append(args, *isRead)
	}
	return filters, args
}

func (r *FeedbackRepo) Create(ctx context.Context, f *models.Feedback) error {
	query := `INSERT INTO feedbacks (user_name, user_phone)
              VALUES ($1, $2)
              RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, f.UserName, f.UserPhone).
		Scan(&f.ID, &f.CreatedAt)
}

func (r *FeedbackRepo) Count(ctx context.Context, phone string, isRead *bool) (int, error) {
	where, args := buildFeedbackFilters(phone, isRead)
	query := `SELECT COUNT(*) FROM feedbacks WHERE ` + where

	var total int
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *FeedbackRepo) List(ctx context.Context, limit, offset int, phone string, isRead *bool) ([]models.Feedback, error) {
	where, args := buildFeedbackFilters(phone, isRead)
	args = append(args, limit, offset)

	query := `SELECT ` + feedbackFields + `
              FROM feedbacks
              WHERE ` + where + `
              ORDER BY created_at DESC
              LIMIT $` + fmt.Sprint(len(args)-1) + ` OFFSET $` + fmt.Sprint(len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Feedback
	for rows.Next() {
		f, err := scanFeedback(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, rows.Err()
}

func (r *FeedbackRepo) MarkAsRead(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE feedbacks SET is_read = true WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *FeedbackRepo) Delete(ctx context.Context, id int) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM feedbacks WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
