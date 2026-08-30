package application

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ConvCursor struct {
	Date time.Time
	ID   uuid.UUID
}

func encodeConvCursor(date time.Time, id uuid.UUID) (string, error) {
	data, err := json.Marshal(ConvCursor{
		Date: date,
		ID:   id,
	})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeConvCursor(value string) (ConvCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return ConvCursor{}, err
	}

	var cursor ConvCursor

	if err := json.Unmarshal(data, &cursor); err != nil {
		return ConvCursor{}, err
	}

	return cursor, nil
}

type MessageCursor struct {
	Date time.Time
	ID   int64
}

func encodeMessageCursor(date time.Time, id int64) (string, error) {
	data, err := json.Marshal(MessageCursor{
		Date: date,
		ID:   id,
	})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeMessageCursor(value string) (MessageCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return MessageCursor{}, err
	}

	var cursor MessageCursor

	if err := json.Unmarshal(data, &cursor); err != nil {
		return MessageCursor{}, err
	}

	return cursor, nil
}