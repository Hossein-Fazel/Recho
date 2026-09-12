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
	ID             int64             `json:"id"`
	ConversationID uuid.UUID         `json:"conversation_id"`
	SenderID       uuid.UUID         `json:"sender_id"`
	Type           model.MessageType `json:"type"`
	Text           *TextMessage      `json:"text,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func ToMessageCreateResponse(msg model.Message) MessageCreateResponse {
	var text *TextMessage
	if msg.Type == model.MessageTypeText {
		text = &TextMessage{
			Content: msg.Text.Content,
		}
	}
	return MessageCreateResponse{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           msg.Type,
		Text:           text,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
	}
}

type MessageEditResponse struct {
	ID             int64             `json:"id"`
	ConversationID uuid.UUID         `json:"conversation_id"`
	SenderID       uuid.UUID         `json:"sender_id"`
	Type           model.MessageType `json:"type"`
	Text           *TextMessage      `json:"text,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func ToMessageEditResponse(msg model.Message) MessageEditResponse {
	var text *TextMessage
	if msg.Type == model.MessageTypeText {
		text = &TextMessage{
			Content: msg.Text.Content,
		}
	}
	return MessageEditResponse{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           msg.Type,
		Text:           text,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
	}
}

type MessageDeleteResponse struct {
	ID             int64     `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
}

type TextMessage struct {
	Content string `json:"content,omitempty"`
}
