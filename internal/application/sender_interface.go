package application

import "github.com/google/uuid"

type Sender interface {
	Send(v any, receivers uuid.UUIDs)
}
