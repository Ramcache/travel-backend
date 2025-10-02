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

type UserRepoI interface {
	GetAll(ctx context.Context) ([]models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, u *models.User) error
	Update(ctx context.Context, u *models.User) error
	UpdatePassword(ctx context.Context, id int, password string) error
	Delete(ctx context.Context, id int) error
}

const userFields = `
	id, email, full_name, avatar, role_id, created_at, updated_at
`

func scanUser(row interface{ Scan(dest ...any) error }, withPassword bool) (models.User, error) {
	var u models.User
	if withPassword {
		err := row.Scan(&u.ID, &u.Email, &u.Password, &u.FullName, &u.Avatar, &u.RoleID, &u.CreatedAt, &u.UpdatedAt)
		return u, err
	}
	err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.Avatar, &u.RoleID, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	query := `SELECT id, email, full_name, role_id, created_at, updated_at FROM users ORDER BY id ASC`
	rows, err := r.db.Query(ctx, query)
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
	return users, rows.Err()
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT ` + userFields + ` FROM users WHERE id=$1`
	u, err := scanUser(r.db.QueryRow(ctx, query, id), false)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password, full_name, avatar, role_id, created_at, updated_at FROM users WHERE email=$1`
	u, err := scanUser(r.db.QueryRow(ctx, query, email), true)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	query := `INSERT INTO users (email, password, full_name, role_id)
              VALUES ($1,$2,$3,$4)
              RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query,
		u.Email, u.Password, u.FullName, u.RoleID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) Update(ctx context.Context, u *models.User) error {
	query := `UPDATE users
              SET full_name=$1, avatar=$2, role_id=$3, updated_at=now()
              WHERE id=$4
              RETURNING updated_at`
	err := r.db.QueryRow(ctx, query,
		u.FullName, u.Avatar, u.RoleID, u.ID,
	).Scan(&u.UpdatedAt)
	if err != nil {
		return mapNotFound(err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int, password string) error {
	query := `UPDATE users SET password=$1, updated_at=now() WHERE id=$2`
	tag, err := r.db.Exec(ctx, query, password, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
