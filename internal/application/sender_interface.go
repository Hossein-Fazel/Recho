package application

import "github.com/google/uuid"



type Sender interface {
	Send(v any, recivers[]uuid.UUID)
}