package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Incomming struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type MessageCreateRequest struct {
	Content        string    `json:"content"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
}
