package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

var (
	ErrUsernameExists      = errors.New("username already exists")
	ErrInvalidLogin        = errors.New("invalid username or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type AuthResult struct {
	User         *model.User
	AccessToken  *model.UserAcccessToken
	RefreshToken *model.UserRefreshToken
}

type AuthService interface {
	Register(ctx context.Context, username string, password string) (*AuthResult, error)
	Login(ctx context.Context, username string, password string) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*model.UserAcccessToken, *model.UserRefreshToken, error)
}

type authService struct {
	users         UserRepo
	at            AccessToken
	rt            RefreshToken
	refreshTokens RefreshTokenRepo
}

func NewAuthService(userRepo UserRepo, at AccessToken, rt RefreshToken, refreshTokens RefreshTokenRepo) AuthService {
	pkg.Logger.Info("Initializing auth service")

	return &authService{
		users:         userRepo,
		at:            at,
		rt:            rt,
		refreshTokens: refreshTokens,
	}
}

func (s *authService) Register(ctx context.Context, username string, password string) (*AuthResult, error) {
	pkg.Logger.Info("Registering user", "username", username)
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
	pkg.Logger.Info("Loging in user", "username", username)
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

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*model.UserAcccessToken, *model.UserRefreshToken, error) {

	tokenHash := s.rt.Hash(refreshToken)

	oldToken, err := s.refreshTokens.GetByHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		return nil, nil, ErrInvalidRefreshToken
	}

	if oldToken.RevokedAt != nil ||
		time.Now().After(oldToken.ExpiresAt) {
		fmt.Println(oldToken.RevokedAt != nil)
		return nil, nil, ErrInvalidRefreshToken
	}

	newToken, userRT, err := s.rt.Generate()
	if err != nil {
		return nil, nil, err
	}

	newToken.UserID = oldToken.UserID

	if err := s.refreshTokens.Rotate(
		ctx,
		oldToken.ID,
		newToken,
	); err != nil {
		return nil, nil, err
	}

	userAT, err := s.at.Generate(
		oldToken.UserID,
	)
	if err != nil {
		return nil, nil, err
	}

	return userAT, userRT, nil
}

func (s *authService) issueTokens(ctx context.Context, userID uuid.UUID) (*model.UserAcccessToken, *model.UserRefreshToken, error) {
	if err := s.refreshTokens.RevokeAllByUser(ctx, userID); err != nil {
		return nil, nil, err
	}

	userAT, err := s.at.Generate(userID)
	if err != nil {
		return nil, nil, err
	}

	rt, userRT, err := s.rt.Generate()
	if err != nil {
		return nil, nil, err
	}
	rt, err = s.refreshTokens.Create(
		ctx,
		userID,
		rt.Hash,
		rt.ExpiresAt,
	)

	if err != nil {
		return nil, nil, err
	}

	return userAT, userRT, nil
}
