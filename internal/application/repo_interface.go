package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/model"
)

type UserRepo interface {
	Create(ctx context.Context, username, passHash string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Exists(ctx context.Context, username string) (bool, error)
}

type RefreshTokenRepo interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.RefreshToken, error)
	GetByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByUser(ctx context.Context, userID uuid.UUID) error
	Rotate(ctx context.Context, oldTokenID uuid.UUID, newToken *model.RefreshToken) error
}

type ConversationRepo interface {
	GetChats(ctx context.Context, userID uuid.UUID, cursorUpdatedAt time.Time, cursorID uuid.UUID, limit int32) ([]*model.UserConversation, error)
}
