package dto

import (
	"encoding/json"

	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type Incomming struct {
	Type      string          `json:"type"`
	RequestID uuid.UUID       `json:"request_id"`
	Payload   json.RawMessage `json:"payload"`
}

type MessageCreateRequest struct {
	Type           model.MessageType `json:"type"`
	Text           *TextMessage      `json:"text,omitempty"`
	ConversationID uuid.UUID         `json:"conversation_id"`
}

type MessageUpdateRequest struct {
	MessageID      int64             `json:"message_id"`
	ConversationID uuid.UUID         `json:"conversation_id"`
	Type           model.MessageType `json:"type"`
	Text           *TextMessage      `json:"text,omitempty"`
}

type MessageDeleteRequest struct {
	MessageID      int64     `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
}
