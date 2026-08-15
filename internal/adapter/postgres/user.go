package postgres_repo

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
)

var (
	URLogger = pkg.Logger.With("component", "userRepo")

	ErrInvalidParams = errors.New("invalid username or password")
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
		return nil, ErrInvalidParams
	}

	user, err := u.sql.CreateUser(ctx, sqlc.CreateUserParams{
		Username: username,
		PasswordHash: passHash,
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
	URLogger.Info("Getting user by username", "username", username)
	user, err := u.sql.GetUserByUsername(ctx, username)

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		PassHash:    user.PasswordHash,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, err
}

func (u *User) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	URLogger.Info("Getting user by id", "id", id)
	user, err := u.sql.GetUserByID(ctx, id)

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		PassHash:    user.PasswordHash,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, err
}

func (u *User) GetForLogin(ctx context.Context, username string) (*model.User, error) {
	URLogger.Info("Getting user for login", "username", username)

	user, err := u.sql.GetUserByUsername(ctx, username)

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		PassHash:    user.PasswordHash,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
		Bio:         user.Bio.String,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, err
}

func (u *User) Exists(ctx context.Context, username string) (bool, error) {
	URLogger.Info("Checking usermae exists", "username", username)

	exists, err := u.sql.UsernameExists(ctx, username)
	return exists, err
}
