package application

import (
	"context"
	"io"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

// Repository

type UpdateUserParams struct {
	ID          uuid.UUID
	Username    string
	DisplayName string
	AvatarKey   string
	Bio         string
}

type UserRepo interface {
	Create(ctx context.Context, username, passHash string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, params UpdateUserParams) (*model.User, error)
	Exists(ctx context.Context, username string) (bool, error)
	Search(ctx context.Context, username string) ([]*model.UserSearch, error)
	UpdateLastSeen(ctx context.Context, userID uuid.UUID) (time.Time, error)
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

type UpdateGroupParams struct {
	GroupID   uuid.UUID
	Name      string
	Bio       string
	AvatarKey string
}

type GroupRepo interface {
	Create(ctx context.Context, params CreateGroupParams) (*model.Group, error)
	GetByID(ctx context.Context, groupID uuid.UUID) (*model.Group, error)
	GetByInviteCode(ctx context.Context, code string) (*model.GroupPreview, error)
	AddMember(ctx context.Context, groupID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	DeleteGroup(ctx context.Context, groupID uuid.UUID) error
	GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (model.GroupMemberRole, error)
	UpdateInviteCode(ctx context.Context, groupID uuid.UUID, newInviteCode string) error
	Update(ctx context.Context, params UpdateGroupParams) (*model.Group, error)
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

// Storage

type Storage interface {
	Upload(
		ctx context.Context,
		file io.Reader,
		size int64,
		key string,
		contentType string,
		public bool,
	) (string, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string) (string, error)
	GetPresignedURL(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
}

// Presence Repo

type PresenceRepository interface {
	Subscribe(subscriberID, targetUserID uuid.UUID) bool
	Unsubscribe(subscriberID, targetUserID uuid.UUID)
	RemoveSubscriber(subscriberID uuid.UUID)

	SetOnline(userID uuid.UUID)
	SetOffline(userID uuid.UUID)

	GetSubscribers(userID uuid.UUID) uuid.UUIDs
}
