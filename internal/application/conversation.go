package application

import (
	"context"
	"errors"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

type ConverasionService struct {
	ConversationRepo ConversationRepo
}

func NewConverasionService(ConversationRepo ConversationRepo) *ConverasionService {
	pkg.Logger.Info().Msg("Initializing Converasion service")

	return &ConverasionService{
		ConversationRepo: ConversationRepo,
	}
}

type GetUserConversationsParams struct {
	UserID          uuid.UUID
	CursorUpdatedAt time.Time
	CursorID        uuid.UUID
	Limit           int32
}

func (c *ConverasionService) GetUserConversations(ctx context.Context, userID uuid.UUID, cursor string, limit int32) ([]*model.UserConversation, string, error) {
	pkg.Logger.Info().
		Str("username", userID.String()).
		Msg("Get users Conversations")

	if userID == uuid.Nil {
		return nil, "", apperr.InvalidInput("conversation service", "user id required", nil)
	}

	cursorItem, err := decodeConvCursor(cursor)
	if err != nil {
		return []*model.UserConversation{}, "", apperr.InvalidInput("conversation service", "invalid cursor", err)
	}

	list, err := c.ConversationRepo.GetConversations(ctx, GetUserConversationsParams{
		UserID:          userID,
		CursorUpdatedAt: cursorItem.Date,
		CursorID:        cursorItem.ID,
		Limit:           limit,
	})

	if err != nil {
		return []*model.UserConversation{}, "", err
	}

	var newCursor string

	if len(list) < int(limit) {
		newCursor = ""
	} else {
		last := list[len(list)-1]
		newCursor, err = encodeConvCursor(
			last.UpdatedAt,
			last.ConversationID,
		)
	}

	return list, newCursor, nil
}

func (c *ConverasionService) GetConversationByID(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) (*model.UserConversation, error) {
	if userID == uuid.Nil || conversationID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation service", "conversation id is required", nil)
	}

	conv, err := c.ConversationRepo.GetConversationByID(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	return conv, nil
}

func (c *ConverasionService) GetOrCreateDC(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error) {
	pkg.Logger.Info().
		Str("user 1", userOneID.String()).
		Str("user 2", userTwoID.String()).
		Msg("Get user Conversation")

	if userOneID == uuid.Nil || userTwoID == uuid.Nil {
		return uuid.Nil, apperr.InvalidInput("conversation service", "user one and two is required", nil)
	}

	if userOneID.String() > userTwoID.String() {
		userTwoID, userOneID = userOneID, userTwoID
	}

	ConversationID, err := c.ConversationRepo.GetDirectConversation(ctx, userOneID, userTwoID)

	if err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Type == apperr.ErrNotFound {
			ConversationID, err := c.ConversationRepo.CreateDirectConversation(ctx, userOneID, userTwoID)
			if err != nil {
				return uuid.Nil, err
			}
			return ConversationID, nil
		}
		return uuid.Nil, err
	}

	return ConversationID, nil
}

type GetConversationMessagesParams struct {
	ConversationID  uuid.UUID
	CursorCreatedAt time.Time
	CursorID        int64
	Limit           int32
}

func (c *ConverasionService) GetConversationMessages(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, cursor string, limit int32) ([]*model.Message, string, error) {
	if conversationID == uuid.Nil {
		return []*model.Message{}, "", apperr.InvalidInput("conversation service", "Conversation id is required", nil)
	}

	isMember, err := c.ConversationRepo.IsConversationMember(ctx, userID, conversationID)
	if err != nil {
		return []*model.Message{}, "", err
	}

	if !isMember {
		return []*model.Message{}, "", apperr.NotFound("conversation service", "Conversation not found", nil)
	}

	CursorItem, err := decodeMessageCursor(cursor)
	if err != nil {
		return []*model.Message{}, "", apperr.InvalidInput("conversation service", "invalid cursor", err)
	}

	list, err := c.ConversationRepo.GetConversationMessages(ctx, GetConversationMessagesParams{
		ConversationID:  conversationID,
		CursorCreatedAt: CursorItem.Date,
		CursorID:        CursorItem.ID,
		Limit:           limit,
	})
	if err != nil {
		return []*model.Message{}, "", err
	}

	var newCursor string

	if len(list) < int(limit) {
		newCursor = ""
	} else {
		last := list[len(list)-1]
		newCursor, err = encodeMessageCursor(
			last.UpdatedAt,
			last.ID,
		)
	}

	return list, newCursor, nil
}

func (c *ConverasionService) GetConversationUsers(ctx context.Context, userID uuid.UUID, convID uuid.UUID) ([]uuid.UUID, error) {
	if convID == uuid.Nil {
		return []uuid.UUID{}, apperr.InvalidInput("conversation service", "Conversation id is required", nil)
	}

	isMember, err := c.ConversationRepo.IsConversationMember(ctx, userID, convID)
	if err != nil {
		return []uuid.UUID{}, err
	}

	if !isMember {
		return []uuid.UUID{}, apperr.NotFound("conversation service", "Conversation not found", nil)
	}

	return  c.ConversationRepo.GetConversationUsers(ctx, convID)
}
