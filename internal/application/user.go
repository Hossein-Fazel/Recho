package application

import (
	"context"
	"strings"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
)

type UserService struct {
	userRepo UserRepo
}

func NewUserService(userRepo UserRepo) *UserService {
	pkg.Logger.Info().Msg("Initializing User service")

	return &UserService{
		userRepo: userRepo,
	}
}

func (u *UserService) Search(ctx context.Context, username string) ([]*model.UserSearch, error) {
	if strings.TrimSpace(username) == "" {
		return []*model.UserSearch{}, apperr.InvalidInput("user service", "username is required", nil)
	}

	return u.userRepo.Search(ctx, username)
}
