package postgres_repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/infra/postgres/sqlc"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
)

type Message struct {
	db  *pgxpool.Pool
	sql *sqlc.Queries
}

var _ application.MessaageRepo = (*Message)(nil)

func NewMessageRepo(sql *sqlc.Queries, db *pgxpool.Pool) *Message {
	pkg.Logger.Info().Msg("Initializing Message Repository")

	return &Message{
		db:  db,
		sql: sql,
	}
}

func (m *Message) Create(ctx context.Context, msg model.Message) (*model.Message, error) {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := m.sql.WithTx(tx)

	switch msg.Type {
	case model.MessageTypeText:
		if msg.Text == nil || msg.Text.Content == "" {
			return nil, apperr.InvalidInput("message repo", "invalid message", nil)
		}

		return m.createTextMessage(ctx, qtx, tx, msg)

	case model.MessageTypeFile:
		if msg.File == nil {
			return nil, apperr.InvalidInput("message repo", "invalid file message", nil)
		}

		return m.createFileMessage(ctx, qtx, tx, msg)

	default:
		return nil, apperr.InvalidInput(
			"message repo",
			"unsupported message type",
			nil,
		)
	}
}

func (m *Message) Update(ctx context.Context, msg model.Message) (*model.Message, error) {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := m.sql.WithTx(tx)

	switch msg.Type {
	case model.MessageTypeText:
		if msg.Text == nil || msg.Text.Content == "" {
			return nil, apperr.InvalidInput("message repo", "invalid message", nil)
		}

		return m.updateTextMessage(ctx, qtx, tx, msg)

	case model.MessageTypeFile:
		if msg.File == nil {
			return nil, apperr.InvalidInput("message repo", "invalid file message", nil)
		}

		return m.updateFileMessage(ctx, qtx, tx, msg)

	default:
		return nil, apperr.InvalidInput("message repo", "unsupported message type", nil)
	}
}

func (m *Message) Delete(ctx context.Context, msg model.Message) error {
	err := m.sql.DeleteMessage(ctx, sqlc.DeleteMessageParams{
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
		UserID:         msg.SenderID,
	})

	if err != nil {
		return apperr.Internal("message repo", err)
	}
	return nil
}

func (m *Message) createTextMessage(ctx context.Context, qtx *sqlc.Queries, tx pgx.Tx, msg model.Message) (*model.Message, error) {
	newMsg, err := qtx.CreateMessage(ctx, sqlc.CreateMessageParams{
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		MessageType:    sqlc.MessageType(msg.Type),
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	msg.ID = newMsg.MessageID
	msg.CreatedAt = newMsg.CreatedAt
	msg.UpdatedAt = newMsg.UpdatedAt

	err = qtx.CreateTextMessage(ctx, sqlc.CreateTextMessageParams{
		ConversationID: msg.ConversationID,
		MessageID:      msg.ID,
		Content:        msg.Text.Content,
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	err = qtx.InsertConversationLastMessage(ctx, sqlc.InsertConversationLastMessageParams{
		CreatedAt: msg.CreatedAt,
		MessageID: pgtype.Int8{
			Int64: msg.ID,
			Valid: msg.ID != 0,
		},
		MessageType: sqlc.NullMessageType{
			MessageType: sqlc.MessageType(msg.Type),
			Valid:       true,
		},
		MessageText: pgtype.Text{
			String: msg.Text.Content,
			Valid:  true,
		},
		FileCategory:   pgtype.Text{},
		ConversationID: msg.ConversationID,
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	return &msg, nil
}

func (m *Message) createFileMessage(ctx context.Context, qtx *sqlc.Queries, tx pgx.Tx, msg model.Message) (*model.Message, error) {
	newMsg, err := qtx.CreateMessage(ctx, sqlc.CreateMessageParams{
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		MessageType:    sqlc.MessageType(msg.Type),
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	msg.ID = newMsg.MessageID
	msg.CreatedAt = newMsg.CreatedAt
	msg.UpdatedAt = newMsg.UpdatedAt

	err = qtx.CreateFileMessage(ctx, sqlc.CreateFileMessageParams{
		ConversationID: msg.ConversationID,
		MessageID:      msg.ID,
		FileKey:        msg.File.Key,
		Category:       msg.File.Category.String(),
		ContentType:    msg.File.ContentType,
		SizeBytes:      msg.File.Size,
		FileName:       msg.File.FileName,
		Caption:        msg.File.Caption,
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	err = qtx.InsertConversationLastMessage(ctx, sqlc.InsertConversationLastMessageParams{
		CreatedAt: msg.CreatedAt,
		MessageID: pgtype.Int8{
			Int64: msg.ID,
			Valid: msg.ID != 0,
		},
		MessageType: sqlc.NullMessageType{
			MessageType: sqlc.MessageType(msg.Type),
			Valid:       true,
		},
		MessageText:    pgtype.Text{String: msg.File.Caption, Valid: msg.File.Caption != ""},
		FileCategory:   pgtype.Text{String: msg.File.Category.String(), Valid: msg.File.Category != ""},
		ConversationID: msg.ConversationID,
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	return &msg, nil
}

func (m *Message) updateTextMessage(ctx context.Context, qtx *sqlc.Queries, tx pgx.Tx, msg model.Message) (*model.Message, error) {
	newTime, err := qtx.UpdateMessage(ctx, sqlc.UpdateMessageParams{
		ConversationID: msg.ConversationID,
		MessageID:      msg.ID,
		UserID:         msg.SenderID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound("message repo", "message not found", err)
		}

		return nil, apperr.Internal("message repo", err)
	}
	msg.UpdatedAt = newTime

	err = qtx.UpdateTextMessage(ctx, sqlc.UpdateTextMessageParams{
		MsgText:        msg.Text.Content,
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
		UserID:         msg.SenderID,
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	err = qtx.UpdateConversationLastMessageContent(ctx, sqlc.UpdateConversationLastMessageContentParams{
		MessageID: pgtype.Int8{
			Int64: msg.ID,
			Valid: msg.ID != 0,
		},
		MessageType: sqlc.NullMessageType{
			MessageType: sqlc.MessageType(msg.Type),
			Valid:       true,
		},
		MessageText: pgtype.Text{
			String: msg.Text.Content,
			Valid:  true,
		},
		FileCategory:   pgtype.Text{},
		ConversationID: msg.ConversationID,
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	return &msg, nil
}

func (m *Message) updateFileMessage(ctx context.Context, qtx *sqlc.Queries, tx pgx.Tx, msg model.Message) (*model.Message, error) {
	newTime, err := qtx.UpdateMessage(ctx, sqlc.UpdateMessageParams{
		ConversationID: msg.ConversationID,
		MessageID:      msg.ID,
		UserID:         msg.SenderID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound("message repo", "message not found", err)
		}
		return nil, apperr.Internal("message repo", err)
	}
	msg.UpdatedAt = newTime

	err = qtx.UpdateFileMessage(ctx, sqlc.UpdateFileMessageParams{
		Caption:        msg.File.Caption,
		ConversationID: msg.ConversationID,
		MessageID:      msg.ID,
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	err = qtx.UpdateConversationLastMessageContent(ctx, sqlc.UpdateConversationLastMessageContentParams{
		MessageID:      pgtype.Int8{Int64: msg.ID, Valid: msg.ID != 0},
		MessageType:    sqlc.NullMessageType{MessageType: sqlc.MessageType(msg.Type), Valid: true},
		MessageText:    pgtype.Text{String: msg.File.Caption, Valid: msg.File.Caption != ""},
		FileCategory:   pgtype.Text{String: msg.File.Category.String(), Valid: msg.File.Category != ""},
		ConversationID: msg.ConversationID,
	})
	if err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperr.Internal("message repo", err)
	}

	return &msg, nil
}
