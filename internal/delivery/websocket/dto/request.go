package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Incomming struct {
	Type      string          `json:"type"`
	RequestID uuid.UUID       `json:"request_id"`
	Payload   json.RawMessage `json:"payload"`
}

type MessageCreateRequest struct {
	Content        string    `json:"content"`
	ConversationID uuid.UUID `json:"conversation_id"`
}

type MessageUpdateRequest struct {
	MessageID      int64     `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Content        string    `json:"content"`
}

type MessageDeleteRequest struct {
	MessageID      int64     `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
}
