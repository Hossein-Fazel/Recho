package postgres_repo

import (
	"context"
	"errors"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

var (
	Alogger = pkg.Logger.With("component", "authRepo")

	ErrInvalidID = errors.New("invalid id")
)

type Auth struct {
	sql *sqlc.Queries
}

func NewAuthRepo(sql *sqlc.Queries) *Auth {
	Alogger.Info("Initializing auth Repository")
	return &Auth{
		sql: sql,
	}
}

func (a *Auth) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.RefreshToken, error) {
	Alogger.Info("Creating refresh token","user id", userID, "tokenHash", tokenHash)
	if userID == uuid.Nil {
		return nil, ErrInvalidID
	}

	rt, err := a.sql.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})

	return &model.RefreshToken{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Hash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		RevokedAt: &rt.RevokedAt.Time,
		CreatedAt: rt.CreatedAt,
	}, err
}

func (a *Auth) GetByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	Alogger.Info("Getting refresh token", "tokenHash", tokenHash)
	rt, err := a.sql.GetRefreshTokenByHash(ctx, tokenHash)

	return &model.RefreshToken{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Hash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		RevokedAt: &rt.RevokedAt.Time,
		CreatedAt: rt.CreatedAt,
	}, err
}

func (a *Auth) RevokeAllByUser(ctx context.Context, userID uuid.UUID) error {
	Alogger.Info("Revoking all refresh tokens", "user id", userID)

	if userID == uuid.Nil {
		return ErrInvalidID
	}

	return a.sql.RevokeAllUserRefreshTokens(ctx, userID)
}

func (a *Auth) Revoke(ctx context.Context, id uuid.UUID) error {
	Alogger.Info("Revoking refresh token", "id", id)

	if id == uuid.Nil {
		return ErrInvalidID
	}

	return a.sql.RevokeRefreshToken(ctx, id)
}
