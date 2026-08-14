package usecase

import (
	"context"
	"errors"

	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

var logger = pkg.Logger.With("component", "auth service")

type UserLogin struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
}

var (
	ErrUsernameExists = errors.New("username already exists")
	ErrInvalidLogin   = errors.New("invalid username or password")
)

type AuthService interface {
	Register(ctx context.Context, username string, password string) (*model.User, error)
	Login(ctx context.Context, username string, password string) (*UserLogin, error)
}

type authService struct {
	users UserRepo
}

func NewAuthService(userrepo UserRepo) AuthService {
	logger.Info("Initializing user service")
	return &authService{
		users: userrepo,
	}
}

func (s *authService) Register(ctx context.Context, username string, password string) (*model.User, error) {
	logger.Info("Registering user", "username", username)
	exists, err := s.users.Exists(ctx, username)

	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrUsernameExists
	}

	hash, err := pkg.HashPassword(password)

	if err != nil {
		return nil, err
	}

	return s.users.Create(
		ctx,
		username,
		hash,
	)
}

func (s *authService) Login(ctx context.Context, username string, password string) (*UserLogin, error) {
	logger.Info("Loging in user", "username", username)
	user, err := s.users.GetForLogin(ctx, username)

	if err != nil {
		return nil, ErrInvalidLogin
	}

	err = pkg.CheckPassword(
		password,
		user.PasswordHash,
	)

	if err != nil {
		return nil, ErrInvalidLogin
	}

	return user, nil
}
