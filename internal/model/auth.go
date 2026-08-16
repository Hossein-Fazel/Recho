package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Hash      string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type UserRefreshToken struct {
	Token string
	TTL   time.Duration
}

type UserAcccessToken struct {
	Token string
	TTL   time.Duration
}
