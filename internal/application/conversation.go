package application

import (
	"context"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

type ConverasionService struct {
	ChatRepo ConversationRepo
}

func NewConverasionService(chatRepo ConversationRepo) *ConverasionService {
	pkg.Logger.Info().Msg("Initializing Converasion service")

	return &ConverasionService{
		ChatRepo: chatRepo,
	}
}

type GetUserChatsParams struct {
	UserID          uuid.UUID
	CursorUpdatedAt time.Time
	CursorID        uuid.UUID
	Limit           int32
}

func (c *ConverasionService) GetUserChats(ctx context.Context, args GetUserChatsParams) ([]*model.UserConversation, error) {
	pkg.Logger.Info().
		Str("username", args.UserID.String()).
		Msg("Get users chats")

	if args.UserID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation service", "user id required", nil)
	}

	return c.ChatRepo.GetChats(ctx, args)
}
