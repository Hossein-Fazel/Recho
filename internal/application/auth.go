package application

import (
	"context"
	"fmt"
	"strings"
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

	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return nil, apperr.InvalidInput("auth service", "username and password are required", nil)
	}

	exists, err := s.users.Exists(ctx, username)

	if err != nil {
		return nil, err
	}

	if exists {
		return nil, apperr.Conflict("auth service", "username already exists", err)
	}

	isValid := pkg.ValidatePassword(password)
	if !isValid {
		return nil, apperr.InvalidInput(
			"auth service",
			"password must be at least 8 characters long and include at least one uppercase letter, lowercase letter, number, and special character",
			nil,
		)
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

	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return nil, apperr.InvalidInput("auth service", "username and password are required", nil)
	}

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

	if strings.TrimSpace(refreshToken) == "" {
		return nil, nil, apperr.InvalidInput("auth service", "refresh token is required", nil)
	}

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
	if userID == uuid.Nil {
		return nil, nil, apperr.InvalidInput("auth service", "user id is required", nil)
	}
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
