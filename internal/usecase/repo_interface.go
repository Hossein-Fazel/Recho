package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/model"
)

type UserRepo interface {
	Create(ctx context.Context, username, passHash string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetForLogin(ctx context.Context, username string) (*UserLogin, error)
	Exists(ctx context.Context, username string) (bool, error)
}