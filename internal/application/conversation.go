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

type ConversationService struct {
	conversationRepo ConversationRepo
	storage          *StorageService
}

func NewConversationService(conversationRepo ConversationRepo, storage *StorageService) *ConversationService {
	pkg.Logger.Info().Msg("Initializing Converasion service")

	return &ConversationService{
		conversationRepo: conversationRepo,
		storage:          storage,
	}
}

type GetUserConversationsParams struct {
	UserID          uuid.UUID
	CursorUpdatedAt time.Time
	CursorID        uuid.UUID
	Limit           int32
}

func (c *ConversationService) GetUserConversations(ctx context.Context, userID uuid.UUID, cursor string, limit int32) ([]*model.UserConversation, string, error) {
	pkg.Logger.Info().
		Str("username", userID.String()).
		Msg("Get users Conversations")

	if userID == uuid.Nil {
		return nil, "", apperr.InvalidInput("conversation service", "user id required", nil)
	}

	cursorItem, err := pkg.DecodeConvCursor(cursor)
	if err != nil {
		return []*model.UserConversation{}, "", apperr.InvalidInput("conversation service", "invalid cursor", err)
	}

	list, err := c.conversationRepo.GetConversations(ctx, GetUserConversationsParams{
		UserID:          userID,
		CursorUpdatedAt: cursorItem.Date,
		CursorID:        cursorItem.ID,
		Limit:           limit,
	})

	if err != nil {
		return []*model.UserConversation{}, "", err
	}

	for idx, item := range list {
		item.AvatarKey = c.fixUrl(ctx, item.AvatarKey)
		item.GroupAvatarKey = c.fixUrl(ctx, item.GroupAvatarKey)

		list[idx] = item
	}

	var newCursor string

	if len(list) < int(limit) {
		newCursor = ""
	} else {
		last := list[len(list)-1]
		newCursor, err = pkg.EncodeConvCursor(
			last.UpdatedAt,
			last.ConversationID,
		)
	}

	return list, newCursor, nil
}

func (c *ConversationService) GetConversationByID(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID) (*model.UserConversation, error) {
	if userID == uuid.Nil || conversationID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation service", "conversation id is required", nil)
	}

	conv, err := c.conversationRepo.GetConversationByID(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	conv.AvatarKey = c.fixUrl(ctx, conv.AvatarKey)
	conv.GroupAvatarKey = c.fixUrl(ctx, conv.GroupAvatarKey)

	return conv, nil
}

func (c *ConversationService) GetOrCreateDC(ctx context.Context, userOneID uuid.UUID, userTwoID uuid.UUID) (uuid.UUID, error) {
	pkg.Logger.Info().
		Str("user 1", userOneID.String()).
		Str("user 2", userTwoID.String()).
		Msg("Get user Conversation")

	if userOneID == uuid.Nil || userTwoID == uuid.Nil {
		return uuid.Nil, apperr.InvalidInput("conversation service", "user one and two is required", nil)
	}

	if userOneID == userTwoID {
		return uuid.Nil, apperr.InvalidInput("conversation service", "user one and two must be different", nil)
	}

	if userOneID.String() > userTwoID.String() {
		userTwoID, userOneID = userOneID, userTwoID
	}

	ConversationID, err := c.conversationRepo.GetDirectConversation(ctx, userOneID, userTwoID)

	if err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Type == apperr.ErrNotFound {
			ConversationID, err := c.conversationRepo.CreateDirectConversation(ctx, userOneID, userTwoID)
			if err != nil {
				return uuid.Nil, err
			}
			return ConversationID, nil
		}
		return uuid.Nil, err
	}

	return ConversationID, nil
}

func (c *ConversationService) UploadMessageMedia(ctx context.Context, userID, conversationID uuid.UUID, category model.MediaCategory, file model.UploadedFile) (*model.Media, error) {
	if userID == uuid.Nil || conversationID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation service", "invalid user or conversation id", nil)
	}

	isMember, err := c.conversationRepo.IsConversationMember(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, apperr.NotFound("conversation service", "Conversation not found", nil)
	}

	return c.storage.UploadMessageMedia(ctx, conversationID, category, file)
}

type GetConversationMessagesParams struct {
	ConversationID  uuid.UUID
	CursorCreatedAt time.Time
	CursorID        int64
	Limit           int32
}

func (c *ConversationService) GetConversationMessages(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, cursor string, limit int32) ([]*model.Message, string, error) {
	if conversationID == uuid.Nil {
		return []*model.Message{}, "", apperr.InvalidInput("conversation service", "Conversation id is required", nil)
	}

	isMember, err := c.conversationRepo.IsConversationMember(ctx, userID, conversationID)
	if err != nil {
		return []*model.Message{}, "", err
	}

	if !isMember {
		return []*model.Message{}, "", apperr.NotFound("conversation service", "Conversation not found", nil)
	}

	CursorItem, err := pkg.DecodeMessageCursor(cursor)
	if err != nil {
		return []*model.Message{}, "", apperr.InvalidInput("conversation service", "invalid cursor", err)
	}

	list, err := c.conversationRepo.GetConversationMessages(ctx, GetConversationMessagesParams{
		ConversationID:  conversationID,
		CursorCreatedAt: CursorItem.Date,
		CursorID:        CursorItem.ID,
		Limit:           limit,
	})
	if err != nil {
		return []*model.Message{}, "", err
	}

	for _, message := range list {
		if message.File != nil && message.File.Key != "" {
			if url, err := c.storage.URL(ctx, message.File.Key); err == nil {
				message.File.URL = url
			}
		}
	}

	var newCursor string

	if len(list) < int(limit) {
		newCursor = ""
	} else {
		last := list[len(list)-1]
		newCursor, err = pkg.EncodeMessageCursor(
			last.UpdatedAt,
			last.ID,
		)
	}

	return list, newCursor, nil
}

func (c *ConversationService) GetConversationUserIDs(ctx context.Context, userID uuid.UUID, convID uuid.UUID) ([]uuid.UUID, error) {
	if convID == uuid.Nil {
		return []uuid.UUID{}, apperr.InvalidInput("conversation service", "Conversation id is required", nil)
	}

	isMember, err := c.conversationRepo.IsConversationMember(ctx, userID, convID)
	if err != nil {
		return []uuid.UUID{}, err
	}

	if !isMember {
		return []uuid.UUID{}, apperr.NotFound("conversation service", "Conversation not found", nil)
	}

	return c.conversationRepo.GetConversationUsers(ctx, convID)
}

type GetGroupMembersParams struct {
	GroupID uuid.UUID
	UserID  uuid.UUID
	Cursor  string
	Limit   int32
}

func (c *ConversationService) GetGroupMembers(ctx context.Context, params GetGroupMembersParams) ([]*model.GroupMember, string, error) {
	if params.GroupID == uuid.Nil {
		return []*model.GroupMember{}, "", apperr.InvalidInput("conversation service", "Conversation id is required", nil)
	}

	isMember, err := c.conversationRepo.IsConversationMember(ctx, params.UserID, params.GroupID)
	if err != nil {
		return []*model.GroupMember{}, "", err
	}

	if !isMember {
		return []*model.GroupMember{}, "", apperr.NotFound("conversation service", "Conversation not found", nil)
	}

	cursor, err := pkg.DecodeGroupMemberCursor(params.Cursor)

	members, err := c.conversationRepo.GetGroupMembers(ctx, params.GroupID, cursor.ID, params.Limit)
	if err != nil {
		return []*model.GroupMember{}, "", err
	}

	for idx, member := range members {
		member.AvatarKey = c.fixUrl(ctx, member.AvatarKey)
		members[idx] = member
	}

	var newCursor string

	if len(members) < int(params.Limit) {
		newCursor = ""
	} else {
		last := members[len(members)-1]
		newCursor, err = pkg.EncodeGroupMemberCursor(
			last.UserID,
		)
	}

	return members, newCursor, nil
}

func (c *ConversationService) GetConversationInfo(ctx context.Context, userID, convID uuid.UUID) (*model.ConversationInfo, error) {
	if convID == uuid.Nil {
		return nil, apperr.InvalidInput("conversation service", "Conversation id is required", nil)
	}

	isMember, err := c.conversationRepo.IsConversationMember(ctx, userID, convID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, apperr.NotFound("conversation service", "Conversation not found", nil)
	}

	info, err := c.conversationRepo.GetInfo(ctx, userID, convID)
	if err != nil {
		return nil, err
	}

	if info.User != nil {
		info.User.AvatarKey = c.fixUrl(ctx, info.User.AvatarKey)
	}

	if info.Group != nil {
		info.Group.GroupAvatarKey = c.fixUrl(ctx, info.Group.GroupAvatarKey)
	}

	return info, nil
}

func (c *ConversationService) fixUrl(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}

	url, err := c.storage.URL(ctx, key)
	if err != nil {
		return ""
	}
	return url
}
