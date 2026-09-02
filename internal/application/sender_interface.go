package application

import (
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type SendItems struct {
	RequestID uuid.UUID
	Event     model.Event
	Content   any
	Recievers uuid.UUIDs
}

type Sender interface {
	Send(items SendItems)
}
