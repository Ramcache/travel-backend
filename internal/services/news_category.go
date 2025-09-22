package services

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

var ErrCategoryNotFound = errors.New("category not found")

type NewsCategoryService struct {
	repo *repository.NewsCategoryRepository
	log  *zap.SugaredLogger
}

func NewNewsCategoryService(repo *repository.NewsCategoryRepository, log *zap.SugaredLogger) *NewsCategoryService {
	return &NewsCategoryService{repo: repo, log: log}
}

func (s *NewsCategoryService) List(ctx context.Context) ([]models.NewsCategory, error) {
	return s.repo.List(ctx)
}

func (s *NewsCategoryService) GetByID(ctx context.Context, id int) (*models.NewsCategory, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCategoryNotFound
	}
	return c, nil
}

func (s *NewsCategoryService) Create(ctx context.Context, req models.CreateNewsCategoryRequest) (*models.NewsCategory, error) {
	if req.Slug == "" || req.Title == "" {
		return nil, helpers.ErrInvalidInput("slug и title обязательны")
	}

	c := &models.NewsCategory{Slug: req.Slug, Title: req.Title}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	s.log.Infow("category_created", "id", c.ID, "slug", c.Slug)
	return c, nil
}

func (s *NewsCategoryService) Update(ctx context.Context, id int, req models.UpdateNewsCategoryRequest) (*models.NewsCategory, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrCategoryNotFound
	}

	if req.Slug != nil {
		c.Slug = *req.Slug
	}
	if req.Title != nil {
		c.Title = *req.Title
	}

	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	s.log.Infow("category_updated", "id", id)
	return c, nil
}

func (s *NewsCategoryService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.log.Infow("category_deleted", "id", id)
	return nil
}
