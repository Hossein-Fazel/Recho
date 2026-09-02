package postgres_repo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
)

type Message struct {
	sql *sqlc.Queries
}

func NewMessageRepo(sql *sqlc.Queries) *Message {
	pkg.Logger.Info().Msg("Initializing Message Repository")

	return &Message{
		sql: sql,
	}
}

func (m *Message) Create(ctx context.Context, msg model.Message) (*model.Message, error) {
	if msg.ConversationID == uuid.Nil || msg.SenderID == uuid.Nil || msg.Content == "" {
		return nil, apperr.InvalidInput("message repo", "invalid message", nil)
	}

	res, err := m.sql.CreateMessage(ctx, sqlc.CreateMessageParams{
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
	})

	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	msg.ID = res.MessageID
	msg.CreatedAt = res.CreatedAt
	msg.UpdatedAt = res.UpdatedAt
	return &msg, nil
}

func (m *Message) Update(ctx context.Context, msg model.Message) (*model.Message, error) {
	if msg.ConversationID == uuid.Nil || msg.SenderID == uuid.Nil || msg.Content == "" {
		return nil, apperr.InvalidInput("message repo", "invalid message", nil)
	}

	newTime, err := m.sql.UpdateMessage(ctx, sqlc.UpdateMessageParams{
		Msgtext:        msg.Content,
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
		UserID:         msg.SenderID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound(
				"message repo",
				"message not found",
				err,
			)
		}

		return nil, apperr.Internal("message repo", err)
	}

	msg.UpdatedAt = newTime
	return &msg, nil
}

func (m *Message) Delete(ctx context.Context, msg model.Message) error {
	if msg.ConversationID == uuid.Nil || msg.SenderID == uuid.Nil {
		return apperr.InvalidInput("message repo", "invalid message", nil)
	}

	err := m.sql.DeleteMessage(ctx, sqlc.DeleteMessageParams{
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
		UserID:         msg.SenderID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NotFound(
				"message repo",
				"message not found",
				err,
			)
		}

		return apperr.Internal("message repo", err)
	}

	return nil
}
