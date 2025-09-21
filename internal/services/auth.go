package services

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type AuthService struct {
	repo      *repository.UserRepository
	jwtSecret string
	log       *zap.SugaredLogger
}

func NewAuthService(repo *repository.UserRepository, jwtSecret string, log *zap.SugaredLogger) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret, log: log}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	hash, err := helpers.HashPassword(req.Password)
	if err != nil {
		s.log.Errorw("hash_password_failed", "err", err)
		return nil, err
	}

	user := &models.User{
		Email:    req.Email,
		Password: hash,
		FullName: req.FullName,
		RoleID:   1,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Warnw("user_create_failed", "email", req.Email, "err", err)
		return nil, err
	}
	s.log.Infow("user_registered", "user_id", user.ID, "email", user.Email)
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (string, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}
	if !helpers.CheckPassword(user.Password, req.Password) {
		return "", errors.New("invalid email or password")
	}

	token, err := helpers.GenerateJWT(s.jwtSecret, user.ID, user.FullName, user.RoleID, 24*time.Hour)
	if err != nil {
		s.log.Errorw("jwt_generate_failed", "user_id", user.ID, "err", err)
		return "", err
	}
	s.log.Infow("user_login", "user_id", user.ID, "role_id", user.RoleID)

	return token, nil
}
