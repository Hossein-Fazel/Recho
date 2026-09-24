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
	storage  *StorageService
}

func NewMessageService(msgRepo MessaageRepo, convRepo ConversationRepo, storage *StorageService) *MessageService {
	return &MessageService{
		msgRepo:  msgRepo,
		convRepo: convRepo,
		storage:  storage,
	}
}

func normalize(msg *model.Message) error {
	if msg.Type == "" {
		msg.Type = model.DefaultMessageType()
	}

	if !msg.Type.IsValid() {
		return apperr.InvalidInput("message service", "unsupported message type", nil)
	}

	switch msg.Type {
	case model.MessageTypeText:
		if msg.Text == nil {
			return apperr.InvalidInput("message service", "message content is required", nil)
		}

		msg.Text.Content = strings.TrimSpace(msg.Text.Content)

		if msg.Text.Content == "" {
			return apperr.InvalidInput("message service", "message content is required", nil)
		}
	case model.MessageTypeFile:
		if msg.File == nil {
			return apperr.InvalidInput("message service", "file is required", nil)
		}

		msg.File.Key = strings.TrimSpace(msg.File.Key)
		msg.File.Caption = strings.TrimSpace(msg.File.Caption)

		if msg.File.Key == "" || !msg.File.Category.IsValid() || msg.File.Category == model.MediaCategoryAvatar {
			return apperr.InvalidInput("message service", "invalid file", nil)
		}
		if msg.File.ContentType == "" || msg.File.Size <= 0 {
			return apperr.InvalidInput("message service", "invalid file metadata", nil)
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

	if msg.Type == model.MessageTypeFile {
		if err := m.storage.ValidateMessageMedia(ctx, msg.ConversationID, *msg.File); err != nil {
			return nil, err
		}
	}

	message, err := m.msgRepo.Create(ctx, msg)

	if err != nil {
		return nil, err
	}

	if message.File != nil {
		if url, err := m.storage.URL(ctx, message.File.Key); err == nil {
			message.File.URL = url
		}
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

	if msg.Type == model.MessageTypeFile {
		if err := m.storage.ValidateMessageMedia(ctx, msg.ConversationID, *msg.File); err != nil {
			return nil, err
		}
	}

	newMsg, err := m.msgRepo.Update(ctx, msg)
	if err != nil {
		return nil, err
	}

	if newMsg.File != nil {
		if url, err := m.storage.URL(ctx, newMsg.File.Key); err == nil {
			newMsg.File.URL = url
		}
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

	if err := m.msgRepo.Delete(ctx, msg); err != nil {
		return err
	}
	
	if msg.Type == model.MessageTypeFile && msg.File != nil {
		if msg.File.Key != "" {
			m.storage.Delete(ctx, msg.File.Key)
		}
	}

	return nil
}
