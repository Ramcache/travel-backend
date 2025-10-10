package services

import (
	"context"
	"fmt"
	"net/url"
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

	// Для туров теперь ссылка с ID и типом тура
	for i := range trips {
		trips[i].Link = fmt.Sprintf("https://web95.tech/trip.html?id=%d&type=%s",
			trips[i].ID,
			url.QueryEscape(trips[i].TripType),
		)
	}

	// Для новостей — ссылка по slug
	for i := range news {
		parts := strings.Split(news[i].Link, "/")
		slug := parts[len(parts)-1]
		news[i].Link = fmt.Sprintf("https://web95.tech/article.html?slug=%s", slug)
	}

	// Объединяем результаты
	results := append(trips, news...)

	// Сортировка по дате (по убыванию)
	sort.Slice(results, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, results[i].Date)
		tj, _ := time.Parse(time.RFC3339, results[j].Date)
		return ti.After(tj)
	})

	return results, nil
}
