package model

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText MessageType = "text"
	MessageTypeFile MessageType = "file"
)

func DefaultMessageType() MessageType {
	return MessageTypeText
}

func (t MessageType) IsValid() bool {
	switch t {
	case MessageTypeText, MessageTypeFile:
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
	File *FileMessage
}

type TextMessage struct {
	Content string
}

type FileMessage struct {
	Key         string
	URL         string
	Category    MediaCategory
	ContentType string
	Size        int64
	FileName    string
	Caption     string
}
