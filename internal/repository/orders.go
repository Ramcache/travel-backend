package repository

import (
	"context"
	"fmt"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{db: db}
}

// Создать заказ
func (r *OrderRepo) Create(ctx context.Context, o *models.Order) error {
	query := `INSERT INTO orders (trip_id, user_name, user_phone, status)
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, o.TripID, o.UserName, o.UserPhone, o.Status).
		Scan(&o.ID, &o.CreatedAt)
}

// Список заказов (для админки)
func (r *OrderRepo) List(ctx context.Context) ([]models.Order, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, trip_id, user_name, user_phone, status, is_read, created_at
        FROM orders
        ORDER BY created_at DESC`)
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
