package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ramcache/travel-backend/internal/models"
)

var ErrNotFound = errors.New("record not found")

type NewsRepository struct {
	db *pgxpool.Pool
}

func NewNewsRepository(db *pgxpool.Pool) *NewsRepository { return &NewsRepository{db: db} }

type NewsFilter struct {
	CategoryID int
	MediaType  string
	Search     string
	Status     string
	Limit      int
	Offset     int
}

// List — список новостей с фильтрацией
func (r *NewsRepository) List(ctx context.Context, f NewsFilter) ([]models.News, int, error) {
	var (
		where []string
		args  []any
		idx   = 1
	)

	if f.Status != "" {
		where = append(where, fmt.Sprintf("n.status=$%d", idx))
		args = append(args, f.Status)
		idx++
	}
	if f.CategoryID > 0 {
		where = append(where, fmt.Sprintf("n.category_id=$%d", idx))
		args = append(args, f.CategoryID)
		idx++
	}
	if f.MediaType != "" {
		where = append(where, fmt.Sprintf("n.media_type=$%d", idx))
		args = append(args, f.MediaType)
		idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(n.title ILIKE $%d OR n.excerpt ILIKE $%d)", idx, idx+1))
		args = append(args, "%"+f.Search+"%", "%"+f.Search+"%")
		idx += 2
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	// count
	var total int
	if err := r.db.QueryRow(ctx, "SELECT count(*) FROM news n "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// data
	args = append(args, f.Limit, f.Offset)
	q := `
SELECT n.id, n.slug, n.title, n.excerpt, n.content,
       n.category_id, c.title AS category, 
       n.media_type, n.preview_url, n.video_url,
       n.comments_count, n.reposts_count, n.views_count,
       n.author_id, n.status, n.published_at,
       n.created_at, n.updated_at
FROM news n
LEFT JOIN news_categories c ON n.category_id = c.id
` + whereSQL + `
ORDER BY n.published_at DESC, n.id DESC
LIMIT $` + fmt.Sprint(idx) + ` OFFSET $` + fmt.Sprint(idx+1)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.News
	for rows.Next() {
		var n models.News
		if err := rows.Scan(
			&n.ID, &n.Slug, &n.Title, &n.Excerpt, &n.Content,
			&n.CategoryID, &n.Category,
			&n.MediaType, &n.PreviewURL, &n.VideoURL,
			&n.CommentsCount, &n.RepostsCount, &n.ViewsCount,
			&n.AuthorID, &n.Status, &n.PublishedAt,
			&n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, rows.Err()
}

func (r *NewsRepository) GetByID(ctx context.Context, id int) (*models.News, error) {
	return r.getOne(ctx, "n.id", id)
}
func (r *NewsRepository) GetBySlug(ctx context.Context, slug string) (*models.News, error) {
	return r.getOne(ctx, "n.slug", slug)
}

func (r *NewsRepository) getOne(ctx context.Context, by string, val any) (*models.News, error) {
	q := `
SELECT n.id, n.slug, n.title, n.excerpt, n.content,
       n.category_id, c.title AS category, 
       n.media_type, n.preview_url, n.video_url,
       n.comments_count, n.reposts_count, n.views_count,
       n.author_id, n.status, n.published_at,
       n.created_at, n.updated_at
FROM news n
LEFT JOIN news_categories c ON n.category_id = c.id
WHERE ` + by + ` = $1`

	var n models.News
	err := r.db.QueryRow(ctx, q, val).Scan(
		&n.ID, &n.Slug, &n.Title, &n.Excerpt, &n.Content,
		&n.CategoryID, &n.Category,
		&n.MediaType, &n.PreviewURL, &n.VideoURL,
		&n.CommentsCount, &n.RepostsCount, &n.ViewsCount,
		&n.AuthorID, &n.Status, &n.PublishedAt,
		&n.CreatedAt, &n.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NewsRepository) Create(ctx context.Context, n *models.News) error {
	return r.db.QueryRow(ctx, `
INSERT INTO news (slug, title, excerpt, content, category_id, media_type, preview_url, video_url,
                  comments_count, reposts_count, views_count, author_id, status, published_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,0,0,0,$9,$10,$11)
RETURNING id, created_at, updated_at`,
		n.Slug, n.Title, n.Excerpt, n.Content, n.CategoryID, n.MediaType, n.PreviewURL, n.VideoURL,
		n.AuthorID, n.Status, n.PublishedAt,
	).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)
}

func (r *NewsRepository) Update(ctx context.Context, n *models.News) error {
	err := r.db.QueryRow(ctx, `
UPDATE news
SET slug=$1, title=$2, excerpt=$3, content=$4, category_id=$5, media_type=$6,
    preview_url=$7, video_url=$8, status=$9, published_at=$10, updated_at=now()
WHERE id=$11
RETURNING updated_at`,
		n.Slug, n.Title, n.Excerpt, n.Content, n.CategoryID, n.MediaType,
		n.PreviewURL, n.VideoURL, n.Status, n.PublishedAt, n.ID,
	).Scan(&n.UpdatedAt)
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	return err
}

func (r *NewsRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM news WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *NewsRepository) ExistsSlug(ctx context.Context, slug string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM news WHERE slug=$1)`, slug).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *NewsRepository) IncrementViews(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `UPDATE news SET views_count = views_count + 1 WHERE id = $1`, id)
	return err
}

func (r *NewsRepository) GetRecent(ctx context.Context, limit int) ([]models.News, error) {
	q := `
SELECT n.id, n.slug, n.title, n.excerpt,
       n.preview_url, n.media_type,
       n.category_id, c.title AS category,
       n.published_at, n.comments_count, n.reposts_count, n.views_count,
       n.created_at, n.updated_at
FROM news n
LEFT JOIN news_categories c ON n.category_id = c.id
WHERE n.status = 'published'
ORDER BY n.published_at DESC, n.id DESC
LIMIT $1`

	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.News
	for rows.Next() {
		var n models.News
		if err := rows.Scan(
			&n.ID, &n.Slug, &n.Title, &n.Excerpt,
			&n.PreviewURL, &n.MediaType,
			&n.CategoryID, &n.Category,
			&n.PublishedAt, &n.CommentsCount, &n.RepostsCount, &n.ViewsCount,
			&n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, rows.Err()
}

func (r *NewsRepository) GetPopular(ctx context.Context, limit int) ([]models.News, error) {
	q := `
SELECT n.id, n.slug, n.title, n.excerpt,
       n.preview_url, n.media_type,
       n.category_id, c.title AS category,
       n.published_at, n.comments_count, n.reposts_count, n.views_count,
       n.created_at, n.updated_at
FROM news n
LEFT JOIN news_categories c ON n.category_id = c.id
WHERE n.status = 'published'
ORDER BY n.views_count DESC, n.published_at DESC
LIMIT $1`

	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.News
	for rows.Next() {
		var n models.News
		if err := rows.Scan(
			&n.ID, &n.Slug, &n.Title, &n.Excerpt,
			&n.PreviewURL, &n.MediaType,
			&n.CategoryID, &n.Category,
			&n.PublishedAt, &n.CommentsCount, &n.RepostsCount, &n.ViewsCount,
			&n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, rows.Err()
}
