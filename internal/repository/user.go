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

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (email, password, full_name) VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, user.Email, user.Password, user.FullName).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password, full_name, created_at, updated_at FROM users WHERE email=$1`
	row := r.db.QueryRow(ctx, query, email)

	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &u.FullName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
