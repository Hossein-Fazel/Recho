package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	Username    string
	PassHash    string
	DisplayName string
	AvatarURL   string
	Bio         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
