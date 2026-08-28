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

type ConverasionService interface {
	GetUserConversations(ctx context.Context, args GetUserConversationsParams) ([]*model.UserConversation, error)
	GetOrCreateDC(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error)
	GetConversationMessages(ctx context.Context, userID uuid.UUID, getConversationparams GetConversationMessagesParams) ([]*model.Message, error)
}

type converasionService struct {
	ConversationRepo ConversationRepo
}

func NewConverasionService(ConversationRepo ConversationRepo) ConverasionService {
	pkg.Logger.Info().Msg("Initializing Converasion service")

	return &converasionService{
		ConversationRepo: ConversationRepo,
	}
}

type GetUserConversationsParams struct {
	UserID          uuid.UUID
	CursorUpdatedAt time.Time
	CursorID        uuid.UUID
	Limit           int32
}

func (c *converasionService) GetUserConversations(ctx context.Context, args GetUserConversationsParams) ([]*model.UserConversation, error) {
	pkg.Logger.Info().
		Str("username", args.UserID.String()).
		Msg("Get users Conversations")

	if args.UserID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation service", "user id required", nil)
	}

	return c.ConversationRepo.GetConversations(ctx, args)
}

func (c *converasionService) GetOrCreateDC(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error) {
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
	ConversationID          uuid.UUID
	CursorCreatedAt time.Time
	CursorID        uuid.UUID
	Limit           int32
}

func (c *converasionService) GetConversationMessages(ctx context.Context, userID uuid.UUID, getConvMessageParams GetConversationMessagesParams) ([]*model.Message, error) {
	if getConvMessageParams.ConversationID == uuid.Nil {
		return []*model.Message{}, apperr.InvalidInput("conversation service", "Conversation id is required", nil)
	}

	isMember, err := c.ConversationRepo.IsConversationMember(ctx, userID, getConvMessageParams.ConversationID)
	if err != nil {
		return []*model.Message{}, err
	}

	if !isMember {
		return []*model.Message{}, apperr.InvalidInput("conversation service", "you don't access to this Conversation", nil)
	}

	list, err := c.ConversationRepo.GetConversationMessages(ctx, getConvMessageParams)
	if err != nil {
		return []*model.Message{}, err
	}

	return list, nil
}
