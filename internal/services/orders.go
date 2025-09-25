package services

import (
	"context"
	"database/sql"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type OrderService struct {
	repo *repository.OrderRepo
}

type OrdersWithTotal struct {
	Total  int            `json:"total"`
	Orders []models.Order `json:"orders"`
}

func NewOrderService(repo *repository.OrderRepo) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) Create(ctx context.Context, tripID int, userName, userPhone string) (*models.Order, error) {
	order := &models.Order{
		TripID: models.NullInt32{
			NullInt32: sql.NullInt32{
				Int32: int32(tripID),
				Valid: tripID > 0, // если 0 → будет NULL
			},
		},
		UserName:  userName,
		UserPhone: userPhone,
		Status:    "new",
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) List(ctx context.Context, limit, offset int, status, phone string, isRead *bool) (*OrdersWithTotal, error) {
	total, err := s.repo.Count(ctx, status, phone, isRead)
	if err != nil {
		return nil, err
	}

	orders, err := s.repo.List(ctx, limit, offset, status, phone, isRead)
	if err != nil {
		return nil, err
	}

	return &OrdersWithTotal{
		Total:  total,
		Orders: orders,
	}, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *OrderService) MarkAsRead(ctx context.Context, id int) error {
	return s.repo.MarkAsRead(ctx, id)
}
