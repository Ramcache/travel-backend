package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Ramcache/travel-backend/internal/models"
)

type OrderRepo struct {
	db DB
}

func NewOrderRepo(db DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, o *models.Order) error {
	var tripID sql.NullInt32

	if o.TripID.Valid {
		tripID = sql.NullInt32{
			Int32: o.TripID.Int32,
			Valid: true,
		}
	} else {
		tripID = sql.NullInt32{Valid: false}
	}

	query := `INSERT INTO orders (trip_id, user_name, user_phone, status)
              VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	return r.db.QueryRow(ctx, query,
		tripID,
		o.UserName,
		o.UserPhone,
		o.Status,
	).Scan(&o.ID, &o.CreatedAt)
}

// Count возвращает количество заказов по тем же фильтрам
func (r *OrderRepo) Count(ctx context.Context, status, phone string, isRead *bool) (int, error) {
	query := `SELECT COUNT(*) FROM orders WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, status)
		argID++
	}
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

// List с пагинацией и фильтрацией
func (r *OrderRepo) List(ctx context.Context, limit, offset int, status, phone string, isRead *bool) ([]models.Order, error) {
	query := `
        SELECT id, trip_id, user_name, user_phone, status, is_read, created_at
        FROM orders
        WHERE 1=1
    `
	args := []interface{}{}
	argID := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, status)
		argID++
	}
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

	var list []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(
			&o.ID,
			&o.TripID,
			&o.UserName,
			&o.UserPhone,
			&o.Status,
			&o.IsRead,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	cmd, err := r.db.Exec(ctx,
		`UPDATE orders SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %d", id)
	}
	return nil
}

// MarkAsRead — отметить заказ как прочитанный
func (r *OrderRepo) MarkAsRead(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `UPDATE orders SET is_read = true WHERE id = $1`, id)
	return err
}

func (r *OrderRepo) Delete(ctx context.Context, id int) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %d", id)
	}
	return nil
}
