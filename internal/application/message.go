package application

import (
	"context"
	"strings"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type MessageService struct {
	msgRepo  MessaageRepo
	convRepo ConversationRepo
}

func NewMessageService(msgRepo MessaageRepo, convRepo ConversationRepo) *MessageService {
	return &MessageService{
		msgRepo:  msgRepo,
		convRepo: convRepo,
	}
}

func normalize(msg *model.Message) error {
	if msg.Type == "" {
		msg.Type = model.DefaultMessageType()
	}

	if !msg.Type.IsValid() {
		return apperr.InvalidInput("message service", "unsupported message type", nil)
	}

	if msg.Type == model.MessageTypeText {
		if msg.Text == nil {
			return apperr.InvalidInput("message service", "message content is required", nil)
		}

		msg.Text.Content = strings.TrimSpace(msg.Text.Content)

		if msg.Text.Content == "" {
			return apperr.InvalidInput("message service", "message content is required", nil)
		}
	}

	if msg.ConversationID == uuid.Nil || msg.SenderID == uuid.Nil {
		return apperr.InvalidInput("message service", "invalid message", nil)
	}

	return nil
}

func (m *MessageService) Create(ctx context.Context, msg model.Message) (*model.Message, error) {
	if err := normalize(&msg); err != nil {
		return nil, err
	}

	isMember, err := m.convRepo.IsConversationMember(ctx, msg.SenderID, msg.ConversationID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, apperr.NotFound("message service", "Conversation not found", nil)
	}

	message, err := m.msgRepo.Create(ctx, msg)

	if err != nil {
		return nil, err
	}
	return message, nil
}

func (m *MessageService) Update(ctx context.Context, msg model.Message) (*model.Message, error) {
	if err := normalize(&msg); err != nil {
		return nil, err
	}

	isMember, err := m.convRepo.IsConversationMember(ctx, msg.SenderID, msg.ConversationID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, apperr.NotFound("message service", "Conversation not found", nil)
	}

	newMsg, err := m.msgRepo.Update(ctx, msg)
	if err != nil {
		return nil, err
	}
	return newMsg, nil
}

func (m *MessageService) Delete(ctx context.Context, msg model.Message) error {
	if msg.ConversationID == uuid.Nil || msg.SenderID == uuid.Nil {
		return apperr.InvalidInput("message repo", "invalid message", nil)
	}

	isMember, err := m.convRepo.IsConversationMember(ctx, msg.SenderID, msg.ConversationID)
	if err != nil {
		return err
	}

	if !isMember {
		return apperr.NotFound("message service", "Conversation not found", nil)
	}

	return m.msgRepo.Delete(ctx, msg)
}
