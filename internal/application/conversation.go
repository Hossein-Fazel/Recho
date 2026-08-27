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
	GetUserChats(ctx context.Context, args GetUserChatsParams) ([]*model.UserConversation, error)
	GetOrCreateDC(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error)
}

type converasionService struct {
	chatRepo ConversationRepo
}

func NewConverasionService(chatRepo ConversationRepo) ConverasionService {
	pkg.Logger.Info().Msg("Initializing Converasion service")

	return &converasionService{
		chatRepo: chatRepo,
	}
}

type GetUserChatsParams struct {
	UserID          uuid.UUID
	CursorUpdatedAt time.Time
	CursorID        uuid.UUID
	Limit           int32
}

func (c *converasionService) GetUserChats(ctx context.Context, args GetUserChatsParams) ([]*model.UserConversation, error) {
	pkg.Logger.Info().
		Str("username", args.UserID.String()).
		Msg("Get users chats")

	if args.UserID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation service", "user id required", nil)
	}

	return c.chatRepo.GetChats(ctx, args)
}

func (c *converasionService) GetOrCreateDC(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error) {
	pkg.Logger.Info().
		Str("user 1", userOneID.String()).
		Str("user 2", userTwoID.String()).
		Msg("Get user chat")

	if userOneID == uuid.Nil || userTwoID == uuid.Nil {
		return uuid.Nil, apperr.InvalidInput("conversation service", "user one and two is required", nil)
	}

	if userOneID.String() > userTwoID.String() {
		userTwoID, userOneID = userOneID, userTwoID
	}

	chatID, err := c.chatRepo.GetDirectConversation(ctx, userOneID, userTwoID)

	if err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Type == apperr.ErrNotFound {
			chatID, err := c.chatRepo.CreateDirectConversation(ctx, userOneID, userTwoID)
			if err != nil {
				return uuid.Nil, err
			}
			return chatID, nil
		}
		return uuid.Nil, err
	}

	return chatID, nil
}
