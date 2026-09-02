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

func (m *MessageService) Create(ctx context.Context, msg model.Message) (*model.Message, error) {
	msg.Content = strings.TrimSpace(msg.Content)
	if msg.ConversationID == uuid.Nil || msg.SenderID == uuid.Nil || msg.Content == "" {
		return nil, apperr.InvalidInput("message repo", "invalid message", nil)
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
	msg.Content = strings.TrimSpace(msg.Content)
	if msg.ConversationID == uuid.Nil || msg.SenderID == uuid.Nil || msg.Content == "" {
		return nil, apperr.InvalidInput("message repo", "invalid message", nil)
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
