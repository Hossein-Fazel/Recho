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

var _ application.ConversationRepo = (*Conversation)(nil)

func NewConversationRepo(sql *sqlc.Queries, db *pgxpool.Pool) *Conversation {
	pkg.Logger.Info().Msg("Initializing Conversation Repository")

	return &Conversation{
		sql: sql,
		db:  db,
	}
}

func (c *Conversation) GetConversations(ctx context.Context, args application.GetUserConversationsParams) ([]*model.UserConversation, error) {
	pkg.Logger.Info().
		Str("user id", args.UserID.String()).
		Str("cursor_update_at", args.CursorUpdatedAt.String()).
		Str("cursorID", args.CursorID.String()).
		Msg("Getting Conversation list")
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

	convList := make([]*model.UserConversation, 0, len(list))
	for _, conv := range list {
		uID, _ := uuid.Parse(conv.UserID.String())
		lmID := conv.LastMessageID.Int64

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
			LastMessageType:      model.MessageType(conv.LastMessageType.MessageType),
			LastMessageText:      conv.LastMessageText.String,
			LastMessageCreatedAt: conv.LastMessageCreatedAt.Time,
			UpdatedAt:            conv.UpdatedAt,
		})
	}

	return convList, nil
}

func (r *Conversation) GetConversationByID(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) (*model.UserConversation, error) {
	conv, err := r.sql.GetConversationByID(
		ctx,
		sqlc.GetConversationByIDParams{
			UserID:         userID,
			ConversationID: conversationID,
		},
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"Conversation repo",
				"conversation not found",
				err,
			)
		}

		return nil, apperr.Internal("Conversation repo", err)
	}

	uID, _ := uuid.Parse(conv.UserID.String())

	return &model.UserConversation{
		ConversationID:       conv.ConversationID,
		ConversationType:     conv.ConversationType,
		UserID:               uID,
		Username:             conv.Username.String,
		DisplayName:          conv.DisplayName.String,
		AvatarUrl:            conv.AvatarUrl.String,
		GroupName:            conv.GroupName.String,
		GroupAvatarUrl:       conv.GroupAvatarUrl.String,
		LastMessageID:        conv.LastMessageID.Int64,
		LastMessageType:      model.MessageType(conv.LastMessageType.MessageType),
		LastMessageText:      conv.LastMessageText.String,
		LastMessageCreatedAt: conv.LastMessageCreatedAt.Time,
		UpdatedAt:            conv.UpdatedAt,
	}, nil
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

	conversationID, err := q.InsertConversation(ctx, sqlc.ConversationTypeDirect)

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

func (r *Conversation) IsConversationMember(ctx context.Context, userID uuid.UUID, ConversationID uuid.UUID) (bool, error) {
	isMember, err := r.sql.IsConversationMember(ctx, sqlc.IsConversationMemberParams{
		ConversationID: ConversationID,
		UserID:         userID,
	})

	if err != nil {
		return false, apperr.Internal("Conversation repo", err)
	}

	return isMember, nil
}

func (r *Conversation) GetConversationMessages(ctx context.Context, params application.GetConversationMessagesParams) ([]*model.Message, error) {
	list, err := r.sql.GetConversationMessages(
		ctx,
		sqlc.GetConversationMessagesParams{
			ConversationID: params.ConversationID,

			CursorCreatedAt: pgtype.Timestamptz{
				Time:  params.CursorCreatedAt,
				Valid: !params.CursorCreatedAt.IsZero(),
			},

			CursorID: pgtype.Int8{
				Int64: params.CursorID,
				Valid: params.CursorID != 0,
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

func (r *Conversation) GetConversationUsers(ctx context.Context, convID uuid.UUID) ([]uuid.UUID, error) {
	uIDs, err := r.sql.GetConvUsers(ctx, convID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"Conversation repo",
				"users not found",
				err,
			)
		}

		return nil, apperr.Internal("Conversation repo", err)
	}

	return uIDs, nil
}

func (r *Conversation) GetGroupMembers(ctx context.Context, GID uuid.UUID, cursorUID uuid.UUID, limit int32) ([]*model.GroupMember, error) {
	members, err := r.sql.GetGroupMembers(ctx, sqlc.GetGroupMembersParams{
		ConversationID: GID,
		CursorUserID:   cursorUID,
		CursorLimit:    limit,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"Conversation repo",
				"users not found",
				err,
			)
		}

		return nil, apperr.Internal("Conversation repo", err)
	}

	var res []*model.GroupMember

	for _, member := range members {
		res = append(res, toModelGroupMember(member))
	}
	return res, nil
}

func (r *Conversation) GetInfo(ctx context.Context, userID uuid.UUID, CID uuid.UUID) (*model.ConversationInfo, error) {
	info, err := r.sql.GetConversationInfo(ctx, sqlc.GetConversationInfoParams{
		ConversationID: CID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"Conversation repo",
				"users not found",
				err,
			)
		}

		return nil, apperr.Internal("Conversation repo", err)
	}

	return toConversationInfo(info), nil
}
