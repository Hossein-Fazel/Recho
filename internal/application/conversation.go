package application

import (
	"context"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

type ConverasionService interface {
	GetUserChats(ctx context.Context, args GetUserChatsParams) ([]*model.UserConversation, error)
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
