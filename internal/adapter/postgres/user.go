package postgres_repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
)

var (
	URLogger = pkg.Logger.With("component", "userRepo")
)

type User struct {
	sql *sqlc.Queries
}

func NewUserRepo(sql *sqlc.Queries) *User {
	URLogger.Info("Initializing User Repository")
	return &User{
		sql: sql,
	}
}

func (u *User) Create(ctx context.Context, username, passHash string) (*model.User, error) {
	URLogger.Info("Creating user", "username", username)
	if username == "" || passHash == "" {
		return nil, apperr.InvalidInput("user repo", "invalid username or password", nil)
	}

	user, err := u.sql.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     username,
		PasswordHash: passHash,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return nil, apperr.Internal("user repo", err)
		}

		switch pgErr.Code {
		case "23505": // unique_violation
			return nil, apperr.Conflict(
				"user repo",
				"re already exists",
				err,
			)
		default:
			return nil, apperr.Internal("user repo", err)
		}
	}

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

func (u *User) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	URLogger.Info("Getting user by username", "username", username)
	user, err := u.sql.GetUserByUsername(ctx, username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"user repo",
				"user not found",
				err,
			)
		}

		return nil, apperr.Internal("user repo", err)
	}

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		PassHash:    user.PasswordHash,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

func (u *User) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	URLogger.Info("Getting user by id", "id", id)
	user, err := u.sql.GetUserByID(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"user repo",
				"user not found",
				err,
			)
		}

		return nil, apperr.Internal("user repo", err)
	}

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		PassHash:    user.PasswordHash,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

func (u *User) Exists(ctx context.Context, username string) (bool, error) {
	URLogger.Info("Checking username existence", "username", username)

	exists, err := u.sql.UsernameExists(ctx, username)
	if err != nil {
		return false, apperr.Internal("user repo", err)
	}

	return exists, nil
}