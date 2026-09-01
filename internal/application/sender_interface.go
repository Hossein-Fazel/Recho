package application

import "github.com/google/uuid"

type Sender interface {
	Send(reqID uuid.UUID, v any, receivers uuid.UUIDs)
}
