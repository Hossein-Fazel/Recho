package postgres_repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
)

type User struct {
	sql *sqlc.Queries
}

var _ application.UserRepo = (*User)(nil)

func NewUserRepo(sql *sqlc.Queries) *User {
	pkg.Logger.Info().Msg("Initializing User Repository")

	return &User{
		sql: sql,
	}
}

func (u *User) Create(ctx context.Context, username, passHash string) (*model.User, error) {
	pkg.Logger.Info().
		Str("username", username).
		Msg("Creating user")

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
				"user already exists",
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
	pkg.Logger.Info().
		Str("username", username).
		Msg("Getting user by username")

	user, err := u.sql.GetUserByUsername(ctx, username)

	if err != nil {
		pkg.Logger.Info().AnErr("err", err)
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
	pkg.Logger.Info().
		Str("id", id.String()).
		Msg("Getting user by id")

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
	pkg.Logger.Info().
		Str("username", username).
		Msg("Checking username existence")

	exists, err := u.sql.UsernameExists(ctx, username)
	if err != nil {
		return false, apperr.Internal("user repo", err)
	}

	return exists, nil
}

func (u *User) Search(ctx context.Context, username string) ([]*model.UserSearch, error) {
	list, err := u.sql.SearchUsers(ctx, pgtype.Text{
		String: username,
		Valid:  true,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*model.UserSearch{}, apperr.NotFound(
				"user repo",
				"no search result",
				err,
			)
		}

		return []*model.UserSearch{}, apperr.Internal("user repo", err)
	}

	res := []*model.UserSearch{}

	for _, user := range list {
		res = append(res, pgUserSearch2modeUserSearch(user))
	}
	return res, nil
}
