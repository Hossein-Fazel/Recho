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
	AvatarKey   string
	Bio         string
	LastSeen    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserSearch struct {
	ID          uuid.UUID
	Username    string
	DisplayName string
	AvatarKey   string
}

type UserStatus struct {
	ID     uuid.UUID
	Online bool
}
