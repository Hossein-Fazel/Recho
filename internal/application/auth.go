package application

import (
	"context"
	"fmt"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

type AuthResult struct {
	User         *model.User
	AccessToken  *model.UserAcccessToken
	RefreshToken *model.UserRefreshToken
}

type AuthService struct {
	users         UserRepo
	at            AccessToken
	rt            RefreshToken
	refreshTokens RefreshTokenRepo
}

func NewAuthService(userRepo UserRepo, at AccessToken, rt RefreshToken, refreshTokens RefreshTokenRepo) *AuthService {
	pkg.Logger.Info().Msg("Initializing auth service")

	return &AuthService{
		users:         userRepo,
		at:            at,
		rt:            rt,
		refreshTokens: refreshTokens,
	}
}

func (s *AuthService) Register(ctx context.Context, username string, password string) (*AuthResult, error) {
	pkg.Logger.Info().
		Str("username", username).
		Msg("Registering user")

	exists, err := s.users.Exists(ctx, username)

	if err != nil {
		return nil, err
	}

	if exists {
		return nil, apperr.Conflict("auth service", "username already exists", err)
	}

	hash, err := pkg.HashPassword(password)

	if err != nil {
		return nil, apperr.Internal("auth service", err)
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
		return nil, apperr.Internal("auth service", err)
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, username string, password string) (*AuthResult, error) {
	pkg.Logger.Info().
		Str("username", username).
		Msg("Loging in user")

	user, err := s.users.GetByUsername(ctx, username)

	if err != nil {
		return nil, err
	}

	err = pkg.CheckPassword(
		password,
		user.PassHash,
	)

	if err != nil {
		return nil, apperr.InvalidInput("auth service", "invalid username or password", err)
	}

	accessToken, refreshToken, err := s.issueTokens(
		ctx,
		user.ID,
	)

	if err != nil {
		return nil, apperr.Internal("auth service", err)
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.UserAcccessToken, *model.UserRefreshToken, error) {

	tokenHash := s.rt.Hash(refreshToken)

	oldToken, err := s.refreshTokens.GetByHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		return nil, nil, err
	}

	if oldToken.RevokedAt != nil ||
		time.Now().After(oldToken.ExpiresAt) {
		fmt.Println(oldToken.RevokedAt != nil)
		return nil, nil, apperr.InvalidInput("auth service", "invalid refresh token", nil)
	}

	newToken, userRT, err := s.rt.Generate()
	if err != nil {
		return nil, nil, apperr.Internal("auth service", err)
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
		return nil, nil, apperr.Internal("auth service", err)
	}

	return userAT, userRT, nil
}

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID) (*model.UserAcccessToken, *model.UserRefreshToken, error) {
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
