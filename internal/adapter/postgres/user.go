package postgres_repo

import (
	"context"
	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/usecase"
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/jackc/pgx/v5/pgtype"
)

var logger = pkg.Logger.With("component", "userRepo")

type User struct {
	sql *sqlc.Queries
}

func NewUserRepo(sql *sqlc.Queries) *User {
	logger.Info("Initializing User Repository")
	return &User{
		sql: sql,
	}
}

func (u *User) Create(ctx context.Context, username, passHash string) (*model.User, error) {
	logger.Info("Creating user", "username", username)
	user, err := u.sql.CreateUser(ctx, sqlc.CreateUserParams{
		Username: username,
		PasswordHash: pgtype.Text{
			String: passHash,
			Valid:  true,
		},
	})

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, err
}

func (u *User) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	logger.Info("Getting user by username", "username", username)
	user, err := u.sql.GetUserByUsername(ctx, username)

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, err
}

func (u *User) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	logger.Info("Getting user by id", "id", id)
	user, err := u.sql.GetUserByID(ctx, id)

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, err
}

func (u *User) GetForLogin(ctx context.Context, username string) (*usecase.UserLogin, error) {
	logger.Info("Getting user for login", "username", username)

	user, err := u.sql.GetUserForLogin(ctx, username)
	
	id, err1 := uuid.Parse(user.ID.String())
	if err == nil && err1 != nil {
		return nil, err1
	}

	return &usecase.UserLogin{
		ID: id,
		Username: user.Username,
		PasswordHash: user.PasswordHash.String,
	}, err
}

func (u *User) Exists(ctx context.Context, username string) (bool, error) {
	logger.Info("Checking usermae exists", "username", username)

	exists, err := u.sql.UsernameExists(ctx, username)
	return exists, err
}
