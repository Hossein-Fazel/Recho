package application

import (
	"context"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

// Repository

type UserRepo interface {
	Create(ctx context.Context, username, passHash string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Exists(ctx context.Context, username string) (bool, error)
	Search(ctx context.Context, username string) ([]*model.UserSearch, error)
}

type RefreshTokenRepo interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.RefreshToken, error)
	GetByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByUser(ctx context.Context, userID uuid.UUID) error
	Rotate(ctx context.Context, oldTokenID uuid.UUID, newToken *model.RefreshToken) error
}

type ConversationRepo interface {
	GetConversations(ctx context.Context, args GetUserConversationsParams) ([]*model.UserConversation, error)
	GetConversationByID(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) (*model.UserConversation, error)
	GetDirectConversation(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error)
	CreateDirectConversation(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error)
	IsConversationMember(ctx context.Context, userID uuid.UUID, ConversationID uuid.UUID) (bool, error)
	GetConversationMessages(ctx context.Context, params GetConversationMessagesParams) ([]*model.Message, error)
	GetConversationUsers(ctx context.Context, convID uuid.UUID) ([]uuid.UUID, error)
	GetInfo(ctx context.Context, userID uuid.UUID, CID uuid.UUID) (*model.ConversationInfo, error)
	GetGroupMembers(ctx context.Context, GID uuid.UUID, cursorUID uuid.UUID, limit int32) ([]*model.GroupMember, error)
}

type GroupRepo interface {
	Create(ctx context.Context, params CreateGroupParams) (*model.Group, error)
	GetByInviteCode(ctx context.Context, code string) (*model.GroupPreview, error)
	AddMember(ctx context.Context, groupID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	DeleteGroup(ctx context.Context, groupID uuid.UUID) error
	GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (model.GroupMemberRole, error)
}

type MessaageRepo interface {
	Create(ctx context.Context, msg model.Message) (*model.Message, error)
	Update(ctx context.Context, msg model.Message) (*model.Message, error)
	Delete(ctx context.Context, msg model.Message) error
}

// Sender

type SendItem struct {
	RequestID uuid.UUID
	Event     model.Event
	Content   any
	Recievers uuid.UUIDs
}

type Sender interface {
	Broadcast(item SendItem)
}

// Token Manager

type AccessToken interface {
	Generate(userID uuid.UUID) (*model.UserAcccessToken, error)
	Validate(token string) (uuid.UUID, error)
}

type RefreshToken interface {
	Generate() (*model.RefreshToken, *model.UserRefreshToken, error)
	Hash(plainText string) string
}
