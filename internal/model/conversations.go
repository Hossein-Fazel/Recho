package model

import (
	"time"

	"github.com/google/uuid"
)

type UserConversation struct {
	ConversationID       uuid.UUID
	ConversationType     string
	UserID               uuid.UUID
	Username             string
	DisplayName          string
	AvatarUrl            string
	GroupName            string
	GroupAvatarUrl       string
	LastMessageID        uuid.UUID
	LastMessageContent   string
	LastMessageCreatedAt time.Time
	UpdatedAt            time.Time
}
