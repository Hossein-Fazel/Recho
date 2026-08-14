package usecase

import 	"github.com/google/uuid"

type UserLogin struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
}
