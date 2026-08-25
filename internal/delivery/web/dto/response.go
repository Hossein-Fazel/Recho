package dto

import (
	"time"

	"github.com/google/uuid"
)

type ErrResponse struct {
	Error  string `json:"error"`
	Type   string `json:"type,omitempty"`
	Module string `json:"module,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type AuthResponse struct {
	Message string `json:"message"`
	User    *User  `json:"user"`
}

type User struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Conversation struct {
	ConversationID       uuid.UUID `json:"conversation_id"`
	ConversationType     string    `json:"conversation_type"`
	UserID               uuid.UUID `json:"user_id"`
	Username             string    `json:"username"`
	DisplayName          string    `json:"display_name"`
	AvatarUrl            string    `json:"avatar_url"`
	GroupName            string    `json:"group_name"`
	GroupAvatarUrl       string    `json:"group_avatar_url"`
	LastMessageID        uuid.UUID `json:"last_message_id"`
	LastMessageContent   string    `json:"last_message_content"`
	LastMessageCreatedAt time.Time `json:"last_message_created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type UserConversations struct {
	Conversations []*Conversation `josn:"Conversations"`
	NextCursor    string          `json:"next_cursor,omitempty"`
}
