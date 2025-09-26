package services

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"time"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrNotFound           = errors.New("user not found")
)

type AuthService struct {
	repo      repository.UserRepoI
	jwtSecret string
	jwtTTL    time.Duration
	log       *zap.SugaredLogger
}

func NewAuthService(repo repository.UserRepoI, jwtSecret string, jwtTTL time.Duration, log *zap.SugaredLogger) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret, jwtTTL: jwtTTL, log: log}
}

type AuthServiceI interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req models.LoginRequest) (string, error)
	UpdateProfile(ctx context.Context, id int, req models.UpdateProfileRequest) (*models.User, error)
	GetByID(ctx context.Context, id int) (*models.User, error)
}

// Register — создаёт нового пользователя
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	// простая проверка
	if req.Email == "" || req.Password == "" {
		return nil, helpers.ErrInvalidInput("email and password are required")
	}

	// проверяем уникальность email
	if existing, _ := s.repo.GetByEmail(ctx, req.Email); existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := helpers.HashPassword(req.Password)
	if err != nil {
		s.log.Errorw("hash_password_failed", "err", err)
		return nil, err
	}

	user := &models.User{
		Email:    req.Email,
		Password: hash,
		FullName: req.FullName,
		RoleID:   1, // default role
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Errorw("user_create_failed", "email", req.Email, "err", err)
		return nil, err
	}

	s.log.Infow("user_registered", "user_id", user.ID, "email", user.Email)
	return user, nil
}

// Login — проверяет креды и выдаёт JWT
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (string, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.log.Warnw("login_failed_user_not_found", "email", req.Email)
		return "", ErrInvalidCredentials
	}
	if !helpers.CheckPassword(user.Password, req.Password) {
		s.log.Warnw("login_failed_invalid_password", "email", req.Email)
		return "", ErrInvalidCredentials
	}

	token, err := helpers.GenerateJWT(s.jwtSecret, user.ID, user.FullName, user.RoleID, s.jwtTTL)
	if err != nil {
		s.log.Errorw("jwt_generate_failed", "user_id", user.ID, "err", err)
		return "", err
	}

	s.log.Infow("user_login", "user_id", user.ID, "role_id", user.RoleID)
	return token, nil
}

// UpdateProfile — обновляет профиль текущего пользователя
func (s *AuthService) UpdateProfile(ctx context.Context, id int, req models.UpdateProfileRequest) (*models.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if req.FullName != nil {
		u.FullName = *req.FullName
	}
	if req.Avatar != nil {
		u.Avatar = req.Avatar
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}

	s.log.Infow("user_profile_updated", "user_id", id)
	return u, nil
}

// GetByID — ищет пользователя по ID
func (s *AuthService) GetByID(ctx context.Context, id int) (*models.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return u, nil
}
