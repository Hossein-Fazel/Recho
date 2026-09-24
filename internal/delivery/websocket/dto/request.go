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
	File           *FileMessage      `json:"file,omitempty"`
	ConversationID uuid.UUID         `json:"conversation_id"`
}

type FileMessage struct {
	Key         string              `json:"key"`
	Category    model.MediaCategory `json:"category"`
	ContentType string              `json:"content_type"`
	Size        int64               `json:"size"`
	FileName    string              `json:"file_name"`
	Caption     string              `json:"caption,omitempty"`
}

type MessageUpdateRequest struct {
	MessageID      int64             `json:"message_id"`
	ConversationID uuid.UUID         `json:"conversation_id"`
	Type           model.MessageType `json:"type"`
	Text           *TextMessage      `json:"text,omitempty"`
	File           *FileMessage      `json:"file,omitempty"`
}

type MessageDeleteRequest struct {
	MessageID      int64     `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
}
