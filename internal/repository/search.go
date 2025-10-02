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

// константы лимитов
const (
	searchLimitTrips = 20
	searchLimitNews  = 20
)

// сканер результата поиска
func scanSearchResultTrip(id int, highlighted string, createdAt time.Time) models.SearchResult {
	return models.SearchResult{
		Type:        "trip",
		ID:          id,
		Title:       highlighted,
		Link:        "/trips/" + strconv.Itoa(id),
		Date:        createdAt.Format(time.RFC3339),
		Highlighted: true,
	}
}

func scanSearchResultNews(id int, highlighted, slug string, createdAt time.Time) models.SearchResult {
	return models.SearchResult{
		Type:        "news",
		ID:          id,
		Title:       highlighted,
		Link:        "/news/" + slug,
		Date:        createdAt.Format(time.RFC3339),
		Highlighted: true,
	}
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
		LIMIT $2
	`, q, searchLimitTrips)
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
		results = append(results, scanSearchResultTrip(id, highlighted, createdAt))
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
		LIMIT $2
	`, q, searchLimitNews)
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
		results = append(results, scanSearchResultNews(id, highlighted, slug, createdAt))
	}
	return results, nil
}
