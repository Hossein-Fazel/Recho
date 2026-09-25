package dto

import (
	"time"

	"github.com/Hossein-Fazel/Recho/internal/model"
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
type MessageMediaUploadResponse struct {
	Key         string              `json:"key"`
	URL         string              `json:"url,omitempty"`
	Category    model.MediaCategory `json:"category"`
	ContentType string              `json:"content_type"`
	Size        int64               `json:"size"`
	FileName    string              `json:"file_name"`
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
	LastSeen    time.Time `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Conversation struct {
	ConversationID       uuid.UUID           `json:"conversation_id"`
	ConversationType     string              `json:"conversation_type"`
	UserID               uuid.UUID           `json:"user_id"`
	Username             string              `json:"username"`
	DisplayName          string              `json:"display_name"`
	AvatarUrl            string              `json:"avatar_url"`
	LastSeen             time.Time           `json:"last_seen"`
	GroupName            string              `json:"group_name"`
	GroupAvatarUrl       string              `json:"group_avatar_url"`
	LastMessageID        int64               `json:"last_message_id"`
	LastMessageType      model.MessageType   `json:"last_message_type,omitempty"`
	LastMessageText      string              `json:"last_message_text"`
	LastFileCategory     model.MediaCategory `json:"last_file_category,omitempty"`
	LastMessageCreatedAt time.Time           `json:"last_message_created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

type UserConversationsResponse struct {
	Conversations []*Conversation `json:"conversations"`
	NextCursor    string          `json:"next_cursor,omitempty"`
}

type UserSearch struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
}

type UserSearchResponse struct {
	Users []*UserSearch `json:"users"`
}

type GetOrCreateDirectConversationResponse struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

type GetConversationMessagesResponse struct {
	Messages   []*Message `json:"messages"`
	NextCursor string     `json:"next_cursor"`
}

type Message struct {
	ID             int64             `json:"id"`
	ConversationID uuid.UUID         `json:"conversation_id"`
	SenderID       uuid.UUID         `json:"sender_id"`
	Type           model.MessageType `json:"type"`
	Text           *TextMessage      `json:"text,omitempty"`
	File           *FileMessage      `json:"file,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type FileMessage struct {
	Key         string              `json:"key"`
	URL         string              `json:"url,omitempty"`
	Category    model.MediaCategory `json:"category"`
	ContentType string              `json:"content_type"`
	Size        int64               `json:"size"`
	FileName    string              `json:"file_name"`
	Caption     string              `json:"caption,omitempty"`
}

type TextMessage struct {
	Content string `json:"content,omitempty"`
}

type ConversationInfo struct {
	ConversationID   uuid.UUID `json:"conversation_id"`
	ConversationType string    `json:"conversation_type"`

	User  *UserInfo  `json:"user,omitempty"`
	Group *GroupInfo `json:"group,omitempty"`
}

type UserInfo struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
}

type GroupInfo struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	AvatarURL   string                `json:"avatar_url"`
	Bio         string                `json:"bio"`
	InviteCode  string                `json:"invite_code,omitempty"`
	Role        model.GroupMemberRole `json:"role,omitempty"`
	MemberCount int64                 `json:"member_count"`
}

type CreateGroupResponse struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	Name           string    `json:"name"`
	Bio            string    `json:"bio"`
	AvatarURL      string    `json:"avatar_url"`
	InviteCode     string    `json:"invite_code"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpdateGroupResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GroupPreviewResponse struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	Name           string    `json:"name"`
	AvatarURL      string    `json:"avatar_url"`
	Bio            string    `json:"bio"`
	MemberCount    int64     `json:"member_count"`
	IsMember       bool      `json:"is_member"`
}

type GetGroupMembersResponse struct {
	Members    []*GroupMember `json:"members"`
	NextCursor string         `json:"next_cursor"`
}

type GroupMember struct {
	UserID      uuid.UUID             `json:"user_id"`
	Username    string                `json:"username"`
	DisplayName string                `json:"display_name"`
	AvatarURL   string                `json:"avatar_url"`
	Role        model.GroupMemberRole `json:"role"`
}

type RotateInviteCodeResponse struct {
	NewInviteCode string `json:"new_invite_code"`
}

type Presence struct {
	UserID uuid.UUID `json:"user_id"`
	Online bool      `json:"online"`
}

type PresenceResponse struct {
	Statuses []Presence `json:"statuses"`
}
