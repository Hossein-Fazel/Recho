package model

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText MessageType = "text"
)

func DefaultMessageType() MessageType {
	return MessageTypeText
}

func (t MessageType) IsValid() bool {
	switch t {
	case MessageTypeText:
		return true
	default:
		return false
	}
}

func (t MessageType) String() string {
	return string(t)
}

type Message struct {
	ID             int64
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Type           MessageType
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Text *TextMessage
}

type TextMessage struct {
	Content string
}
