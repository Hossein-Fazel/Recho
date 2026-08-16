package usecase

import (
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type AccessToken interface {
	Generate(userID uuid.UUID) (*model.UserAcccessToken, error)
	Validate(token string) (uuid.UUID, error)
}

type RefreshToken interface {
	Generate() (*model.RefreshToken, *model.UserRefreshToken, error)
}