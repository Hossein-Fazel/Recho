package api

import (
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
)

const (
	CtxUserID = "user_id"
)

func conv2convRes(conv *model.UserConversation) *dto.Conversation {
	return &dto.Conversation{
		ConversationID:       conv.ConversationID,
		ConversationType:     conv.ConversationType,
		UserID:               conv.UserID,
		Username:             conv.Username,
		DisplayName:          conv.DisplayName,
		AvatarUrl:            conv.AvatarKey,
		GroupName:            conv.GroupName,
		GroupAvatarUrl:       conv.GroupAvatarKey,
		LastMessageID:        conv.LastMessageID,
		LastMessageType:      conv.LastMessageType,
		LastMessageText:      conv.LastMessageText,
		LastMessageCreatedAt: conv.LastMessageCreatedAt,
		UpdatedAt:            conv.UpdatedAt,
	}
}

func convList2convListRes(convs []*model.UserConversation, next string) dto.UserConversationsResponse {
	var res dto.UserConversationsResponse

	res.NextCursor = next
	for _, conv := range convs {
		res.Conversations = append(res.Conversations, conv2convRes(conv))
	}

	return res
}

func group2res(group *model.Group) dto.UpdateGroupResponse {
	return dto.UpdateGroupResponse{
		ID:        group.ID,
		Name:      group.Name,
		AvatarURL: group.AvatarKey,
		Bio:       group.Bio,
		UpdatedAt: group.UpdatedAt,
	}
}

func groupPreview2res(preview *model.GroupPreview) dto.GroupPreviewResponse {
	return dto.GroupPreviewResponse{
		ConversationID: preview.GroupID,
		Name:           preview.Name,
		AvatarURL:      preview.AvatarKey,
		Bio:            preview.Bio,
		MemberCount:    preview.MemberCount,
		IsMember:       preview.IsMember,
	}
}

func uSearch2dtoUSearch(user *model.UserSearch) *dto.UserSearch {
	return &dto.UserSearch{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarKey,
	}
}

func user2dtoUser(user *model.User) *dto.User {
	return &dto.User{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarKey,
		Bio:         user.Bio,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
