package usecase

import (
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type AccessToken interface {
	Generate(userID uuid.UUID) (string, error)
	Validate(token string) (uuid.UUID, error)
}

type RefreshToken interface {
	Generate() (*model.RefreshToken, string, error)
}