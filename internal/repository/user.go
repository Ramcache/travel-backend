package repository

import (
	"context"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, email, full_name, role_id, created_at, updated_at
        FROM users
        ORDER BY id ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.RoleID, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int, password string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password=$1, updated_at=now() WHERE id=$2`,
		password, id,
	)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `SELECT id, email, full_name, role_id, created_at, updated_at FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Email, &u.FullName, &u.RoleID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO users (email, password, full_name, role_id)
         VALUES ($1,$2,$3,$4) RETURNING id, created_at, updated_at`,
		u.Email, u.Password, u.FullName, u.RoleID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) Update(ctx context.Context, u *models.User) error {
	return r.db.QueryRow(ctx,
		`UPDATE users SET full_name=$1, role_id=$2, updated_at=now()
         WHERE id=$3 RETURNING updated_at`,
		u.FullName, u.RoleID, u.ID,
	).Scan(&u.UpdatedAt)
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password, full_name, role_id, created_at, updated_at
         FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Email, &u.Password, &u.FullName, &u.RoleID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
