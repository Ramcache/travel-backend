package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
	"go.uber.org/zap"
)

type FeedbackService struct {
	repo     *repository.FeedbackRepo
	telegram *helpers.TelegramClient
	log      *zap.SugaredLogger
}

func NewFeedbackService(repo *repository.FeedbackRepo, telegram *helpers.TelegramClient, log *zap.SugaredLogger) *FeedbackService {
	return &FeedbackService{repo: repo, telegram: telegram, log: log}
}

func (s *FeedbackService) Create(ctx context.Context, req models.FeedbackRequest) error {
	fb := models.Feedback{
		UserName:  req.UserName,
		UserPhone: req.UserPhone,
	}

	if err := s.repo.Create(ctx, &fb); err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"💬 <b>Новая заявка на консультацию!</b>\n\n"+
			"📅 <b>Дата:</b> %s\n"+
			"👤 <b>Имя:</b> %s\n"+
			"📞 <b>Телефон:</b> <a href=\"tel:%s\">%s</a>",
		time.Now().Format("02.01.2006 15:04"),
		fb.UserName,
		fb.UserPhone, fb.UserPhone,
	)

	if s.telegram != nil {
		if err := s.telegram.SendMessage(msg); err != nil {
			s.log.Errorw("Ошибка отправки feedback в Telegram", "err", err)
			return err
		}
	}

	return nil
}

type FeedbacksWithTotal struct {
	Total     int               `json:"total"`
	Feedbacks []models.Feedback `json:"feedbacks"`
}

func (s *FeedbackService) List(ctx context.Context, limit, offset int, phone string, isRead *bool) (*FeedbacksWithTotal, error) {
	total, err := s.repo.Count(ctx, phone, isRead)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.List(ctx, limit, offset, phone, isRead)
	if err != nil {
		return nil, err
	}

	return &FeedbacksWithTotal{
		Total:     total,
		Feedbacks: items,
	}, nil
}

func (s *FeedbackService) MarkAsRead(ctx context.Context, id int) error {
	return s.repo.MarkAsRead(ctx, id)
}
