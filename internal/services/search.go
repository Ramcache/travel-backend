package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type SearchService struct {
	repo        *repository.SearchRepository
	frontendURL string
}

func NewSearchService(repo *repository.SearchRepository, frontendURL string) *SearchService {
	return &SearchService{repo: repo, frontendURL: frontendURL}
}

func (s *SearchService) GlobalSearch(ctx context.Context, query string) ([]models.SearchResult, error) {
	trips, err := s.repo.SearchTrips(ctx, query)
	if err != nil {
		return nil, err
	}
	news, err := s.repo.SearchNews(ctx, query)
	if err != nil {
		return nil, err
	}

	// проставляем абсолютные ссылки с /api/v1
	for i := range trips {
		trips[i].Link = fmt.Sprintf("%s/api/v1/trips/%d", s.frontendURL, trips[i].ID)
	}
	for i := range news {
		// у news.Link внутри repo уже "/news/{slug}"
		// поэтому вырезаем slug и собираем абсолютный путь
		parts := strings.Split(news[i].Link, "/")
		slug := parts[len(parts)-1]
		news[i].Link = fmt.Sprintf("https://web95.tech/article.html?slug=%s", slug)
	}

	// объединяем
	results := append(trips, news...)

	// сортировка по дате
	sort.Slice(results, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, results[i].Date)
		tj, _ := time.Parse(time.RFC3339, results[j].Date)
		return ti.After(tj)
	})

	return results, nil
}
