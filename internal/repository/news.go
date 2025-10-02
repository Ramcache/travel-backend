package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NewsRepository struct {
	db *pgxpool.Pool
}

func NewNewsRepository(db *pgxpool.Pool) *NewsRepository { return &NewsRepository{db: db} }

// общий SELECT список
const newsFields = `
	n.id, n.slug, n.title, n.excerpt, n.content,
    n.category_id, n.media_type, n.preview_url, n.video_url,
    n.comments_count, n.reposts_count, n.views_count,
    n.author_id, n.status, n.published_at,
    n.created_at, n.updated_at
`

// сканер новости
func scanNews(row interface{ Scan(dest ...any) error }) (models.News, error) {
	var n models.News
	var categoryID sql.NullInt32
	err := row.Scan(
		&n.ID, &n.Slug, &n.Title, &n.Excerpt, &n.Content,
		&categoryID, &n.MediaType, &n.PreviewURL, &n.VideoURL,
		&n.CommentsCount, &n.RepostsCount, &n.ViewsCount,
		&n.AuthorID, &n.Status, &n.PublishedAt,
		&n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return n, err
	}
	if categoryID.Valid {
		val := int(categoryID.Int32)
		n.CategoryID = &val
	}
	return n, nil
}

type NewsFilter struct {
	CategoryID int
	MediaType  string
	Search     string
	Status     string
	Limit      int
	Offset     int
}

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

	// total count
	var total int
	if err := r.db.QueryRow(ctx, "SELECT count(*) FROM news n "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// data query
	args = append(args, f.Limit, f.Offset)
	query := `SELECT ` + newsFields + `
              FROM news n
              ` + whereSQL + `
              ORDER BY n.published_at DESC, n.id DESC
              LIMIT $` + fmt.Sprint(idx) + ` OFFSET $` + fmt.Sprint(idx+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.News
	for rows.Next() {
		n, err := scanNews(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, rows.Err()
}

func (r *NewsRepository) GetByID(ctx context.Context, id int) (*models.News, error) {
	n, err := scanNews(r.db.QueryRow(ctx, `SELECT `+newsFields+` FROM news n WHERE n.id=$1`, id))
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &n, nil
}

func (r *NewsRepository) GetBySlug(ctx context.Context, slug string) (*models.News, error) {
	n, err := scanNews(r.db.QueryRow(ctx, `SELECT `+newsFields+` FROM news n WHERE n.slug=$1`, slug))
	if err != nil {
		return nil, mapNotFound(err)
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
    preview_url=$7, video_url=$8, status=$9, published_at=$10
WHERE id=$11
RETURNING updated_at`,
		n.Slug, n.Title, n.Excerpt, n.Content, n.CategoryID, n.MediaType,
		n.PreviewURL, n.VideoURL, n.Status, n.PublishedAt, n.ID,
	).Scan(&n.UpdatedAt)
	return mapNotFound(err)
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
	query := `SELECT n.id, n.slug, n.title, n.excerpt,
                     n.preview_url, n.media_type,
                     n.category_id, n.published_at,
                     n.comments_count, n.reposts_count, n.views_count,
                     n.created_at, n.updated_at
              FROM news n
              WHERE n.status = 'published'
              ORDER BY n.published_at DESC, n.id DESC
              LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.News
	for rows.Next() {
		n, err := scanNews(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, rows.Err()
}

func (r *NewsRepository) GetPopular(ctx context.Context, limit int) ([]models.News, error) {
	query := `SELECT n.id, n.slug, n.title, n.excerpt,
                     n.preview_url, n.media_type,
                     n.category_id, n.published_at,
                     n.comments_count, n.reposts_count, n.views_count,
                     n.created_at, n.updated_at
              FROM news n
              WHERE n.status = 'published'
              ORDER BY n.views_count DESC, n.published_at DESC
              LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.News
	for rows.Next() {
		n, err := scanNews(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, rows.Err()
}
