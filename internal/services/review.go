package services

import (
	"context"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
	"go.uber.org/zap"
)

type ReviewService struct {
	repo *repository.ReviewRepo
	log  *zap.SugaredLogger
}

func NewReviewService(repo *repository.ReviewRepo, log *zap.SugaredLogger) *ReviewService {
	return &ReviewService{repo: repo, log: log}
}

func (s *ReviewService) Create(ctx context.Context, req models.CreateReviewRequest) (*models.TripReview, error) {
	rev := &models.TripReview{
		TripID:   req.TripID,
		UserName: req.UserName,
		Rating:   req.Rating,
		Comment:  req.Comment,
	}
	if err := s.repo.Create(ctx, rev); err != nil {
		s.log.Errorw("review_create_failed", "trip_id", req.TripID, "err", err)
		return nil, err
	}
	return rev, nil
}

func (s *ReviewService) ListByTrip(ctx context.Context, tripID, limit, offset int) ([]models.TripReview, int, error) {
	return s.repo.ListByTrip(ctx, tripID, limit, offset)
}
