package postgres_repo

import (
	"context"
	"errors"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		return nil, apperr.InvalidInput("token repo", "invalid id", nil)
	}

	rt, err := a.sql.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return nil, apperr.Internal("token repo", err)
		}

		switch pgErr.Code {
		case "23505": // unique_violation
			return nil, apperr.Conflict(
				"token repo",
				"token already exists",
				err,
			)
		case "23503":
			return nil, apperr.InvalidInput(
				"token repo",
				"invalid user reference",
				err,
			)
		default:
			return nil, apperr.Internal("token repo", err)
		}
	}

	return &model.RefreshToken{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Hash:      rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		RevokedAt: nil,
		CreatedAt: rt.CreatedAt,
	}, nil
}

func (a *Token) GetByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	pkg.Logger.Info("Getting refresh token", "tokenHash", tokenHash)
	rt, err := a.sql.GetRefreshTokenByHash(ctx, tokenHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"token repo",
				"token not found",
				err,
			)
		}

		return nil, apperr.Internal("token repo", err)
	}

	var revokeTime *time.Time = &rt.RevokedAt.Time

	if !rt.RevokedAt.Valid {
		revokeTime = nil
	}

	return &model.RefreshToken{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Hash:      rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		RevokedAt: revokeTime,
		CreatedAt: rt.CreatedAt,
	}, nil
}

func (a *Token) RevokeAllByUser(ctx context.Context, userID uuid.UUID) error {
	pkg.Logger.Info("Revoking all refresh tokens", "user id", userID)

	if userID == uuid.Nil {
		return apperr.InvalidInput("token repo", "invalid id", nil)
	}

	return a.sql.RevokeAllUserRefreshTokens(ctx, userID)
}

func (a *Token) Revoke(ctx context.Context, id uuid.UUID) error {
	pkg.Logger.Info("Revoking refresh token", "id", id)

	if id == uuid.Nil {
		return apperr.InvalidInput("token repo", "invalid id", nil)
	}

	err := a.sql.RevokeRefreshToken(ctx, id)
	if err != nil {
		return apperr.Internal("token repo", nil)
	}

	return nil
}

func (r *Token) Rotate(ctx context.Context, oldTokenID uuid.UUID, newToken *model.RefreshToken) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperr.Internal("token repo", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.sql.WithTx(tx)

	if err := qtx.RevokeRefreshToken(ctx, oldTokenID); err != nil {
		return apperr.Internal("token repo", nil)
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
		return apperr.Internal("token repo", nil)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return apperr.Internal("token repo", nil)
	}

	return nil
}
