package postgres_repo

import (
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
)

func pgUserSearch2modeUserSearch(user sqlc.SearchUsersRow) *model.UserSearch {
	return &model.UserSearch{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName.String,
		AvatarURL:   user.AvatarUrl.String,
	}
}

func toModelMessage(msg sqlc.Message) *model.Message {
	return &model.Message{
		ID:             msg.MessageID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
	}
}
