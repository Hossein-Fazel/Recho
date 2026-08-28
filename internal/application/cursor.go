package application

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CursorItem struct {
	Date time.Time
	ID   uuid.UUID
}

func encodeCursor(date time.Time, id uuid.UUID) (string, error) {
	data, err := json.Marshal(CursorItem{
		Date: date,
		ID:   id,
	})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeCursor(value string) (CursorItem, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return CursorItem{}, err
	}

	var cursor CursorItem

	if err := json.Unmarshal(data, &cursor); err != nil {
		return CursorItem{}, err
	}

	return cursor, nil
}
