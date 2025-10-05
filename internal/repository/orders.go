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

const orderFields = `
	id, trip_id, name, date, price, user_name, user_phone, status, is_read, created_at
`

// приватный сканер
func scanOrder(row interface{ Scan(dest ...any) error }) (models.Order, error) {
	var o models.Order
	var name, date, price sql.NullString

	err := row.Scan(
		&o.ID,
		&o.TripID,
		&name,
		&date,
		&price,
		&o.UserName,
		&o.UserPhone,
		&o.Status,
		&o.IsRead,
		&o.CreatedAt,
	)
	if err != nil {
		return o, err
	}

	if name.Valid {
		o.Name = &name.String
	}
	if date.Valid {
		o.Date = &date.String
	}
	if price.Valid {
		o.Price = &price.String
	}

	return o, nil
}

// buildOrderFilters собирает WHERE + args
func buildOrderFilters(status, phone string, isRead *bool) (string, []any) {
	filters := "1=1"
	args := []any{}
	i := 1

	if status != "" {
		filters += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, status)
		i++
	}
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

func (r *OrderRepo) Create(ctx context.Context, o *models.Order) error {
	query := `INSERT INTO orders (trip_id, name, date, price, user_name, user_phone, status)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          RETURNING id, created_at`

	trip := sql.NullInt32{Int32: o.TripID.Int32, Valid: o.TripID.Valid}

	return r.db.QueryRow(ctx, query,
		trip,
		o.Name,
		o.Date,
		o.Price,
		o.UserName,
		o.UserPhone,
		o.Status,
	).Scan(&o.ID, &o.CreatedAt)
}

func (r *OrderRepo) Count(ctx context.Context, status, phone string, isRead *bool) (int, error) {
	where, args := buildOrderFilters(status, phone, isRead)
	query := `SELECT COUNT(*) FROM orders WHERE ` + where

	var total int
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *OrderRepo) List(ctx context.Context, limit, offset int, status, phone string, isRead *bool) ([]models.Order, error) {
	where, args := buildOrderFilters(status, phone, isRead)
	args = append(args, limit, offset)

	query := `SELECT ` + orderFields + `
              FROM orders
              WHERE ` + where + `
              ORDER BY created_at DESC
              LIMIT $` + fmt.Sprint(len(args)-1) + ` OFFSET $` + fmt.Sprint(len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	cmd, err := r.db.Exec(ctx, `UPDATE orders SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *OrderRepo) MarkAsRead(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `UPDATE orders SET is_read = true WHERE id = $1`, id)
	return err
}

func (r *OrderRepo) Delete(ctx context.Context, id int) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM orders WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
