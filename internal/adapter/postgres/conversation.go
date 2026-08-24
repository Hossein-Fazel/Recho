package postgres_repo

import (
	"context"
	"errors"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Conversation struct {
	sql *sqlc.Queries
}

func NewConversationRepo(sql *sqlc.Queries) *Conversation {
	pkg.Logger.Info().Msg("Initializing Conversation Repository")

	return &Conversation{
		sql: sql,
	}
}

func (c *Conversation) GetChats(ctx context.Context, args application.GetUserChatsParams) ([]*model.UserConversation, error) {
	pkg.Logger.Info().
		Str("user id", args.UserID.String()).
		Str("cursor_update_at", args.CursorUpdatedAt.String()).
		Str("cursorID", args.CursorID.String()).
		Msg("Getting chat list")

	if args.UserID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation repo", "invalid id", nil)
	}

	list, err := c.sql.GetUserConversations(
		ctx,
		sqlc.GetUserConversationsParams{
			UserID: args.UserID,

			CursorUpdatedAt: pgtype.Timestamptz{
				Time:  args.CursorUpdatedAt,
				Valid: !args.CursorUpdatedAt.IsZero(),
			},

			CursorID: pgtype.UUID{
				Bytes: args.CursorID,
				Valid: args.CursorID != uuid.Nil,
			},

			QueryLimit: args.Limit,
		},
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"Conversation repo",
				"conversations not found",
				err,
			)
		}

		return nil, apperr.Internal("Conversation repo", err)
	}

	convList := make([]*model.UserConversation, len(list))
	for _, conv := range list {
		uID, _ := uuid.Parse(conv.UserID.String())
		lmID, _ := uuid.Parse(conv.LastMessageID.String())

		convList = append(convList, &model.UserConversation{
			ConversationID:       conv.ConversationID,
			ConversationType:     conv.ConversationType,
			UserID:               uID,
			Username:             conv.Username.String,
			DisplayName:          conv.DisplayName.String,
			AvatarUrl:            conv.AvatarUrl.String,
			GroupName:            conv.GroupName.String,
			GroupAvatarUrl:       conv.GroupAvatarUrl.String,
			LastMessageID:        lmID,
			LastMessageContent:   conv.LastMessageContent.String,
			LastMessageCreatedAt: conv.LastMessageCreatedAt.Time,
		})
	}

	return convList, nil
}
