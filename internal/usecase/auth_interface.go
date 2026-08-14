package usecase

import "github.com/google/uuid"

type AccessToken interface {
	Generate(userID uuid.UUID) (string, error)
	Validate(token string) (uuid.UUID, error)
}

type RefreshToken interface {
	Generate() (plain string, hash string, err error)
}