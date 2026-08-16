package postgres_repo

import (
	"context"
	"errors"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidID = errors.New("invalid id")
)

type Token struct {
	db  *pgxpool.Pool
	sql *sqlc.Queries
}

func NewRefreshTokenRepo(sql *sqlc.Queries, db *pgxpool.Pool) *Token {
	pkg.Logger.Info("Initializing Token Repository")
	return &Token{
		db:  db,
		sql: sql,
	}
}

func (a *Token) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.RefreshToken, error) {
	pkg.Logger.Info("Creating refresh token", "user id", userID, "tokenHash", tokenHash)
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
		Hash:      rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		RevokedAt: &rt.RevokedAt.Time,
		CreatedAt: rt.CreatedAt,
	}, err
}

func (a *Token) GetByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	pkg.Logger.Info("Getting refresh token", "tokenHash", tokenHash)
	rt, err := a.sql.GetRefreshTokenByHash(ctx, tokenHash)

	return &model.RefreshToken{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Hash:      rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		RevokedAt: &rt.RevokedAt.Time,
		CreatedAt: rt.CreatedAt,
	}, err
}

func (a *Token) RevokeAllByUser(ctx context.Context, userID uuid.UUID) error {
	pkg.Logger.Info("Revoking all refresh tokens", "user id", userID)

	if userID == uuid.Nil {
		return ErrInvalidID
	}

	return a.sql.RevokeAllUserRefreshTokens(ctx, userID)
}

func (a *Token) Revoke(ctx context.Context, id uuid.UUID) error {
	pkg.Logger.Info("Revoking refresh token", "id", id)

	if id == uuid.Nil {
		return ErrInvalidID
	}

	return a.sql.RevokeRefreshToken(ctx, id)
}

func (r *Token) Rotate(ctx context.Context, oldTokenID uuid.UUID, newToken *model.RefreshToken) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.sql.WithTx(tx)

	if err := qtx.RevokeRefreshToken(ctx, oldTokenID); err != nil {
		return err
	}

	_, err = qtx.CreateRefreshToken(
		ctx,
		sqlc.CreateRefreshTokenParams{
			UserID:    newToken.UserID,
			TokenHash: newToken.Hash,
			ExpiresAt: newToken.ExpiresAt,
		},
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
