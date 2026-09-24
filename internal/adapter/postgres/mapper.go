package postgres_repo

import (
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func nullText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func pgUserSearch2modeUserSearch(user sqlc.SearchUsersRow) *model.UserSearch {
	return &model.UserSearch{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
		AvatarKey:   user.AvatarKey.String,
	}
}

func toModelMessage(msg sqlc.GetConversationMessagesRow) *model.Message {
	result := &model.Message{
		ID:             msg.MessageID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Type:           model.MessageType(msg.Type),
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
	}

	if msg.Type == sqlc.MessageTypeText {
		result.Text = &model.TextMessage{Content: msg.Content.String}
	}

	if msg.Type == sqlc.MessageTypeFile && msg.FileKey.Valid {
		result.File = &model.FileMessage{
			Key:         msg.FileKey.String,
			Category:    model.MediaCategory(msg.Category.String),
			ContentType: msg.ContentType.String,
			Size:        msg.SizeBytes.Int64,
			FileName:    msg.FileName.String,
			Caption:     msg.Caption.String,
		}
	}

	return result
}

func toModelGroupMember(member sqlc.GetGroupMembersRow) *model.GroupMember {
	return &model.GroupMember{
		UserID:      member.ID,
		Username:    member.Username,
		DisplayName: member.DisplayName.String,
		AvatarKey:   member.AvatarKey.String,
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
			AvatarKey:   row.AvatarKey.String,
			Bio:         row.Bio.String,
		}
	}

	if row.GroupID.Valid {
		group := &model.GroupInfo{
			GroupID:        uuid.UUID(row.GroupID.Bytes),
			GroupName:      row.GroupName.String,
			GroupAvatarKey: row.GroupAvatarKey.String,
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
		AvatarKey:  group.AvatarKey.String,
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
		AvatarKey:   row.AvatarKey.String,
		Bio:         row.Bio.String,
		MemberCount: row.MemberCount,
	}
}
