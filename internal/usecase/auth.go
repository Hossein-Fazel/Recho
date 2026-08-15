package usecase

import (
	"context"
	"errors"

	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

var (
	logger = pkg.Logger.With("component", "auth service")

	ErrUsernameExists = errors.New("username already exists")
	ErrInvalidLogin   = errors.New("invalid username or password")
)

type AuthResult struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

type AuthService interface {
	Register(ctx context.Context, username string, password string) (*AuthResult, error)
	Login(ctx context.Context, username string, password string) (*AuthResult, error)
}

type authService struct {
	users         UserRepo
	at            AccessToken
	rt            RefreshToken
	refreshTokens RefreshTokenRepo
}

func NewAuthService(userRepo UserRepo, at AccessToken, rt RefreshToken, refreshTokens RefreshTokenRepo) AuthService {
	logger.Info("Initializing auth service")

	return &authService{
		users:         userRepo,
		at:            at,
		rt:            rt,
		refreshTokens: refreshTokens,
	}
}

func (s *authService) Register(ctx context.Context, username string, password string) (*AuthResult, error) {
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

	user, err := s.users.Create(
		ctx,
		username,
		hash,
	)

	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.issueTokens(
		ctx,
		user.ID,
	)

	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Login(ctx context.Context, username string, password string) (*AuthResult, error) {
	logger.Info("Loging in user", "username", username)
	user, err := s.users.GetByUsername(ctx, username)

	if err != nil {
		return nil, ErrInvalidLogin
	}

	err = pkg.CheckPassword(
		password,
		user.PassHash,
	)

	if err != nil {
		return nil, ErrInvalidLogin
	}

	accessToken, refreshToken, err := s.issueTokens(
		ctx,
		user.ID,
	)

	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) issueTokens(ctx context.Context, userID uuid.UUID) (string, string, error) {
	if err := s.refreshTokens.RevokeAllByUser(ctx, userID); err != nil {
		return "", "", err
	}

	accessToken, err := s.at.Generate(userID)
	if err != nil {
		return "", "", err
	}

	rt, plainRT, err := s.rt.Generate()
	if err != nil {
		return "", "", err
	}
	rt, err = s.refreshTokens.Create(
		ctx,
		userID,
		rt.Hash,
		rt.ExpiresAt,
	)

	if err != nil {
		return "", "", err
	}

	return accessToken, plainRT, nil
}
