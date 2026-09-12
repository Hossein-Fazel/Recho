package postgres_repo

import (
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

func pgUserSearch2modeUserSearch(user sqlc.SearchUsersRow) *model.UserSearch {
	return &model.UserSearch{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
	}
}

func toModelMessage(msg sqlc.GetConversationMessagesRow) *model.Message {
	var text *model.TextMessage
	if msg.Type == sqlc.MessageTypeText {
		text = &model.TextMessage{
			Content: msg.Content.String,
		}
	}

	return &model.Message{
		ID:             msg.MessageID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           model.MessageType(msg.Type),
		Text:           text,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
	}
}

func toModelGroupMember(member sqlc.GetGroupMembersRow) *model.GroupMember {
	return &model.GroupMember{
		UserID:      member.ID,
		Username:    member.Username,
		DisplayName: member.DisplayName.String,
		AvatarURL:   member.AvatarUrl.String,
		Role:        model.GroupMemberRole(member.MemberRole),
	}
}

func toConversationInfo(row sqlc.GetConversationInfoRow) *model.ConversationInfo {
	info := model.ConversationInfo{
		ID:               row.ID,
		ConversationType: row.ConversationType,
	}

	if row.UserID.Valid {
		info.User = &model.UserInfo{
			UserID:      uuid.UUID(row.UserID.Bytes),
			Username:    row.Username.String,
			DisplayName: row.DisplayName.String,
			AvatarUrl:   row.AvatarUrl.String,
			Bio:         row.Bio.String,
		}
	}

	if row.GroupID.Valid {
		group := &model.GroupInfo{
			GroupID:        uuid.UUID(row.GroupID.Bytes),
			GroupName:      row.GroupName.String,
			GroupAvatarUrl: row.GroupAvatarUrl.String,
			GroupBio:       row.GroupBio.String,
			InviteCode:     row.InviteCode.String,
			MemberCount:    row.MemberCount,
		}

		if row.ViewerRole.Valid {
			group.ViewerRole = model.GroupMemberRole(row.ViewerRole.GroupMemberRole)
		}

		info.Group = group
	}

	return &info
}

func toModelGroup(group sqlc.Group) *model.Group {
	return &model.Group{
		ID:         group.ConversationID,
		Name:       group.Name,
		AvatarURL:  group.AvatarUrl.String,
		Bio:        group.Bio.String,
		InviteCode: group.InviteCode.String,
		CreatedBy:  group.CreatedBy,
		CreatedAt:  group.CreatedAt,
		UpdatedAt:  group.UpdatedAt,
	}
}

func toModelGroupPreview(row sqlc.GetGroupByInviteCodeRow) *model.GroupPreview {
	return &model.GroupPreview{
		GroupID:     row.ConversationID,
		Name:        row.Name,
		AvatarURL:   row.AvatarUrl.String,
		Bio:         row.Bio.String,
		MemberCount: row.MemberCount,
	}
}
