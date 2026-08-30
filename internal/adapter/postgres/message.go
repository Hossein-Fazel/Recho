package postgres_repo

import (
	"context"
	"strings"

	"github.com/google/uuid"

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
	msg.Content = strings.TrimSpace(msg.Content)
	if msg.ConversationID == uuid.Nil || msg.SenderID == uuid.Nil || msg.Content == "" {
		return nil, apperr.InvalidInput("message repo", "invalid message", nil)
	}

	id, err := m.sql.CreateMessage(ctx, sqlc.CreateMessageParams{
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
	})

	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	msg.ID = id
	return &msg, nil
}
