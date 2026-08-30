package dto

import (
	"github.com/google/uuid"
)

type Incomming struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type MessageCreateRequest struct {
	Content        string    `json:"content"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
}
