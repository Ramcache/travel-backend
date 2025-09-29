package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/Ramcache/travel-backend/internal/models"
)

type SearchRepository struct {
	db DB
}

func NewSearchRepository(db DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) SearchTrips(ctx context.Context, q string) ([]models.SearchResult, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id,
		       ts_headline('simple', title, plainto_tsquery(unaccent($1)),
		                   'StartSel=<mark>, StopSel=</mark>, MaxWords=15, MinWords=5, ShortWord=2, HighlightAll=TRUE') AS highlighted,
		       created_at
		FROM trips
		WHERE to_tsvector('simple', title || ' ' || coalesce(description,'')) @@ plainto_tsquery(unaccent($1))
		   OR similarity(unaccent(title), unaccent($1)) > 0.3
		ORDER BY created_at DESC
		LIMIT 20
	`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var id int
		var highlighted string
		var createdAt time.Time
		if err := rows.Scan(&id, &highlighted, &createdAt); err != nil {
			return nil, err
		}
		results = append(results, models.SearchResult{
			Type:        "trip",
			ID:          id,
			Title:       highlighted,
			Link:        "/trips/" + strconv.Itoa(id),
			Date:        createdAt.Format(time.RFC3339),
			Highlighted: true,
		})
	}
	return results, nil
}

func (r *SearchRepository) SearchNews(ctx context.Context, q string) ([]models.SearchResult, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id,
		       ts_headline('simple', title, plainto_tsquery(unaccent($1)),
		                   'StartSel=<mark>, StopSel=</mark>, MaxWords=15, MinWords=5, ShortWord=2, HighlightAll=TRUE') AS highlighted,
		       slug,
		       created_at
		FROM news
		WHERE to_tsvector('simple', unaccent(title || ' ' || coalesce(content,''))) @@ plainto_tsquery(unaccent($1))
		   OR similarity(unaccent(title), unaccent($1)) > 0.3
		ORDER BY created_at DESC
		LIMIT 20
	`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var id int
		var highlighted, slug string
		var createdAt time.Time
		if err := rows.Scan(&id, &highlighted, &slug, &createdAt); err != nil {
			return nil, err
		}
		results = append(results, models.SearchResult{
			Type:        "news",
			ID:          id,
			Title:       highlighted,
			Link:        "/news/" + slug,
			Date:        createdAt.Format(time.RFC3339),
			Highlighted: true,
		})

	}
	return results, nil
}
