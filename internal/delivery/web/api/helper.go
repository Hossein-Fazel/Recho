package api

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

const (
	CtxUserID = "user_id"
)

type conversationCursor struct {
	UpdatedAt time.Time
	ID        uuid.UUID
}

func encodeCursor(updatedAt time.Time, id uuid.UUID) (string, error) {
	data, err := json.Marshal(conversationCursor{
		UpdatedAt: updatedAt,
		ID:        id,
	})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeCursor(value string) (conversationCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return conversationCursor{}, err
	}

	var cursor conversationCursor

	if err := json.Unmarshal(data, &cursor); err != nil {
		return conversationCursor{}, err
	}

	return cursor, nil
}

func convList2convListRes(convs []*model.UserConversation, next string) dto.UserConversationsResponse {
	var res dto.UserConversationsResponse

	res.NextCursor = next
	for _, conv := range convs {
		res.Conversations = append(res.Conversations, &dto.Conversation{
			ConversationID:       conv.ConversationID,
			ConversationType:     conv.ConversationType,
			UserID:               conv.UserID,
			Username:             conv.Username,
			DisplayName:          conv.DisplayName,
			AvatarUrl:            conv.AvatarUrl,
			GroupName:            conv.GroupName,
			GroupAvatarUrl:       conv.GroupAvatarUrl,
			LastMessageID:        conv.LastMessageID,
			LastMessageContent:   conv.LastMessageContent,
			LastMessageCreatedAt: conv.LastMessageCreatedAt,
			UpdatedAt:            conv.UpdatedAt,
		})
	}

	return res
}
