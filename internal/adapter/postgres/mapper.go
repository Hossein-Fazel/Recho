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
