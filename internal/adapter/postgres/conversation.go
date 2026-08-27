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
	"github.com/jackc/pgx/v5/pgxpool"
)

type Conversation struct {
	sql *sqlc.Queries
	db  *pgxpool.Pool
}

func NewConversationRepo(sql *sqlc.Queries, db *pgxpool.Pool) *Conversation {
	pkg.Logger.Info().Msg("Initializing Conversation Repository")

	return &Conversation{
		sql: sql,
		db:  db,
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

func (r *Conversation) GetDirectConversation(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error) {
	conversationID, err := r.sql.GetDirectConversation(ctx, sqlc.GetDirectConversationParams{
		UserOneID: userOneID,
		UserTwoID: userTwoID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound(
				"Conversation repo",
				"conversation not found",
				err,
			)
		}

		return uuid.Nil, apperr.Internal("Conversation repo", err)
	}

	return conversationID, nil
}

func (r *Conversation) CreateDirectConversation(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error) {
	if userOneID == userTwoID {
		return uuid.Nil, apperr.InvalidInput("Conversation repo", "cannot create conversation with yourself", nil)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, apperr.Internal("Conversation repo", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	q := r.sql.WithTx(tx)

	conversationID, err := q.InsertConversation(ctx)

	if err != nil {
		return uuid.Nil, apperr.Internal("Conversation repo", err)
	}

	_, err = q.InsertDirectConversation(
		ctx,
		sqlc.InsertDirectConversationParams{
			ConversationID: conversationID,
			UserOneID:      userOneID,
			UserTwoID:      userTwoID,
		},
	)

	if err != nil {
		return uuid.Nil, apperr.Internal("Conversation repo", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.Nil, apperr.Internal("Conversation repo", err)
	}

	return conversationID, nil
}

func(r *Conversation) IsChatMember(ctx context.Context, userID uuid.UUID, chatID uuid.UUID) (bool, error) {
	isMember, err := r.sql.IsConversationMember(ctx, sqlc.IsConversationMemberParams{
		ConversationID: chatID,
		UserID:         userID,
	})

	if err != nil {
		return false, apperr.Internal("Conversation repo", err)
	}

	return isMember, nil
}

func(r *Conversation) GetChatMessages(ctx context.Context, params application.GetChatMessagesParams) ([]*model.Message, error) {
	if params.ChatID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation repo", "invalid chat id", nil)
	}

	list, err := r.sql.GetConversationMessages(
		ctx,
		sqlc.GetConversationMessagesParams{
			ConversationID: params.ChatID,

			CursorCreatedAt: pgtype.Timestamptz{
				Time:  params.CursorCreatedAt,
				Valid: !params.CursorCreatedAt.IsZero(),
			},

			CursorID: pgtype.UUID{
				Bytes: params.CursorID,
				Valid: params.CursorID != uuid.Nil,
			},

			QueryLimit: params.Limit,
		},
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"Conversation repo",
				"messages not found",
				err,
			)
		}

		return nil, apperr.Internal("Conversation repo", err)
	}

	var res []*model.Message

	for _, msg := range list {
		res = append(res, toModelMessage(msg))
	}

	return res, nil
}
