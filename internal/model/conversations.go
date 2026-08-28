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
	LastMessageID        int64
	LastMessageContent   string
	LastMessageCreatedAt time.Time
	UpdatedAt            time.Time
}
