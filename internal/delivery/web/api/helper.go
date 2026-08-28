package api

import (
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
)

const (
	CtxUserID = "user_id"
)

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

func uSearch2dtoUSearch(user *model.UserSearch) *dto.UserSearch {
	return &dto.UserSearch{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
	}
}
