package repository

import (
	"context"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCategoryNotFound = pgx.ErrNoRows

type NewsCategoryRepository struct {
	db *pgxpool.Pool
}

func NewNewsCategoryRepository(db *pgxpool.Pool) *NewsCategoryRepository {
	return &NewsCategoryRepository{db: db}
}

// общий SELECT список
const newsCategoryFields = `
	id, slug, title, created_at, updated_at
`

// приватный сканер
func scanNewsCategory(row interface{ Scan(dest ...any) error }) (models.NewsCategory, error) {
	var c models.NewsCategory
	err := row.Scan(&c.ID, &c.Slug, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *NewsCategoryRepository) List(ctx context.Context) ([]models.NewsCategory, error) {
	query := `SELECT ` + newsCategoryFields + ` FROM news_categories ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.NewsCategory
	for rows.Next() {
		c, err := scanNewsCategory(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *NewsCategoryRepository) GetByID(ctx context.Context, id int) (*models.NewsCategory, error) {
	c, err := scanNewsCategory(r.db.QueryRow(ctx, `SELECT `+newsCategoryFields+` FROM news_categories WHERE id=$1`, id))
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &c, nil
}

func (r *NewsCategoryRepository) Create(ctx context.Context, c *models.NewsCategory) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO news_categories (slug, title) VALUES ($1,$2) RETURNING id, created_at, updated_at`,
		c.Slug, c.Title).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *NewsCategoryRepository) Update(ctx context.Context, c *models.NewsCategory) error {
	err := r.db.QueryRow(ctx,
		`UPDATE news_categories SET slug=$1, title=$2, updated_at=now() WHERE id=$3 RETURNING updated_at`,
		c.Slug, c.Title, c.ID,
	).Scan(&c.UpdatedAt)
	if err != nil {
		return mapNotFound(err)
	}
	return nil
}

func (r *NewsCategoryRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM news_categories WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *NewsCategoryRepository) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM news_categories WHERE id=$1)`, id).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
