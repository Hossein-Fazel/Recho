package dto

import (
	"time"

	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type WSResponse struct {
	Type      string    `json:"type"`
	RequestID uuid.UUID `json:"request_id"`
	Data      any       `json:"data"`
}
type MessageCreateResponse struct {
	ID             int64                `json:"id"`
	ConversationID uuid.UUID            `json:"conversation_id"`
	SenderID       uuid.UUID            `json:"sender_id"`
	Type           model.MessageType    `json:"type"`
	Text           *TextMessage         `json:"text,omitempty"`
	File           *FileMessageResponse `json:"file,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

func ToMessageCreateResponse(msg model.Message) MessageCreateResponse {
	var text *TextMessage
	if msg.Type == model.MessageTypeText && msg.Text != nil {
		text = &TextMessage{Content: msg.Text.Content}
	}
	var file *FileMessageResponse
	if msg.Type == model.MessageTypeFile && msg.File != nil {
		file = toFileMessageResponse(msg.File)
	}
	return MessageCreateResponse{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           msg.Type,
		Text:           text,
		File:           file,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
	}
}

type MessageEditResponse struct {
	ID             int64                `json:"id"`
	ConversationID uuid.UUID            `json:"conversation_id"`
	SenderID       uuid.UUID            `json:"sender_id"`
	Type           model.MessageType    `json:"type"`
	Text           *TextMessage         `json:"text,omitempty"`
	File           *FileMessageResponse `json:"file,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

func ToMessageEditResponse(msg model.Message) MessageEditResponse {
	var text *TextMessage
	if msg.Type == model.MessageTypeText && msg.Text != nil {
		text = &TextMessage{
			Content: msg.Text.Content,
		}
	}
	var file *FileMessageResponse
	if msg.Type == model.MessageTypeFile && msg.File != nil {
		file = toFileMessageResponse(msg.File)
	}
	return MessageEditResponse{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           msg.Type,
		Text:           text,
		File:           file,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
	}
}

type MessageDeleteResponse struct {
	ID             int64     `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
}

type ConversationDeleteResponse struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

type GroupUpdateResponse struct {
	GroupID   uuid.UUID `json:"group_id"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
}

type PresenceUpdateResponse struct {
	UserID   uuid.UUID  `json:"user_id"`
	Online   bool       `json:"online"`
	LastSeen *time.Time `json:"last_seen,omitempty"`
}

func ToPresenceUpdateResponse(p model.UserPresence) PresenceUpdateResponse {
	return PresenceUpdateResponse{
		UserID:   p.UserID,
		Online:   p.Online,
		LastSeen: p.LastSeen,
	}
}

type TextMessage struct {
	Content string `json:"content,omitempty"`
}
type FileMessageResponse struct {
	Key         string              `json:"key"`
	URL         string              `json:"url,omitempty"`
	Category    model.MediaCategory `json:"category"`
	ContentType string              `json:"content_type"`
	Size        int64               `json:"size"`
	FileName    string              `json:"file_name"`
	Caption     string              `json:"caption,omitempty"`
}

func toFileMessageResponse(file *model.FileMessage) *FileMessageResponse {
	return &FileMessageResponse{
		Key:         file.Key,
		URL:         file.URL,
		Category:    file.Category,
		ContentType: file.ContentType,
		Size:        file.Size,
		FileName:    file.FileName,
		Caption:     file.Caption,
	}
}
