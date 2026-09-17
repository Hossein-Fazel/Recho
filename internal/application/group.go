package application

import (
	"context"
	"errors"
	"strings"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

const (
	maxGroupNameLength = 100
	maxGroupBioLength  = 250
	inviteCodeAttempts = 5
)

type GroupService struct {
	groupRepo        GroupRepo
	conversationRepo ConversationRepo
	storage          *StorageService
	sender           Sender
}

func NewGroupService(
	groupRepo GroupRepo,
	conversationRepo ConversationRepo,
	storage *StorageService,
	sender Sender,
) *GroupService {
	pkg.Logger.Info().Msg("Initializing Group service")

	return &GroupService{
		groupRepo:        groupRepo,
		conversationRepo: conversationRepo,
		storage:          storage,
		sender:           sender,
	}
}

type CreateGroupParams struct {
	Name       string
	Bio        string
	InviteCode string
	CreatedBy  uuid.UUID
}

type UpdateGroupProfileParams struct {
	Name *string
	Bio  *string
}

func (s *GroupService) Create(ctx context.Context, createdBy uuid.UUID, name, bio string) (*model.Group, error) {
	pkg.Logger.Info().
		Str("user id", createdBy.String()).
		Msg("Creating group")

	if createdBy == uuid.Nil {
		return nil, apperr.InvalidInput("group service", "user id required", nil)
	}

	name = strings.TrimSpace(name)
	bio = strings.TrimSpace(bio)

	if name == "" {
		return nil, apperr.InvalidInput("group service", "group name is required", nil)
	}

	if len([]rune(name)) > maxGroupNameLength {
		return nil, apperr.InvalidInput("group service", "group name is too long", nil)
	}

	if len([]rune(bio)) > maxGroupBioLength {
		return nil, apperr.InvalidInput("group service", "group bio is too long", nil)
	}

	var lastErr error

	for range inviteCodeAttempts {
		code, err := pkg.NewInviteCode()
		if err != nil {
			return nil, apperr.Internal("group service", err)
		}

		group, err := s.groupRepo.Create(ctx, CreateGroupParams{
			Name:       name,
			Bio:        bio,
			InviteCode: code,
			CreatedBy:  createdBy,
		})
		if err == nil {
			return group, nil
		}

		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Type == apperr.ErrConflict {
			lastErr = err
			continue
		}

		return nil, err
	}

	return nil, apperr.Internal("group service", lastErr)
}

func (s *GroupService) PreviewByInviteCode(ctx context.Context, userID uuid.UUID, code string) (*model.GroupPreview, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return nil, apperr.InvalidInput("group service", "invite code is required", nil)
	}

	preview, err := s.groupRepo.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if userID != uuid.Nil {
		role, err := s.groupRepo.GetMemberRole(ctx, preview.GroupID, userID)
		if err != nil {
			return nil, err
		}

		preview.IsMember = role != ""
	}

	return preview, nil
}

func (s *GroupService) JoinByInviteCode(ctx context.Context, userID uuid.UUID, code string) (*model.GroupPreview, error) {
	pkg.Logger.Info().
		Str("user id", userID.String()).
		Msg("Joining group by invite code")

	if userID == uuid.Nil {
		return nil, apperr.InvalidInput("group service", "user id required", nil)
	}

	preview, err := s.PreviewByInviteCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}

	if preview.IsMember {
		return preview, nil
	}

	if err := s.groupRepo.AddMember(ctx, preview.GroupID, userID); err != nil {
		return nil, err
	}

	preview.IsMember = true
	preview.MemberCount++

	return preview, nil
}

func (s *GroupService) LeaveGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	pkg.Logger.Info().
		Str("user id", userID.String()).
		Str("group id", groupID.String()).
		Msg("Leaving group")

	if userID == uuid.Nil || groupID == uuid.Nil {
		return apperr.InvalidInput("group service", "user and group id are required", nil)
	}

	isMember, err := s.conversationRepo.IsConversationMember(ctx, userID, groupID)
	if err != nil {
		return err
	}

	if !isMember {
		return apperr.InvalidInput("group service", "invalid group", nil)
	}

	if err := s.groupRepo.RemoveMember(ctx, groupID, userID); err != nil {
		return err
	}

	return nil
}

func (s *GroupService) Delete(ctx context.Context, userID, groupID uuid.UUID) error {
	pkg.Logger.Info().
		Str("user id", userID.String()).
		Str("group id", groupID.String()).
		Msg("request to delete group")

	if userID == uuid.Nil || groupID == uuid.Nil {
		return apperr.InvalidInput("group service", "user and group id are required", nil)
	}

	if err := s.requireOwner(ctx, groupID, userID); err != nil {
		return err
	}

	members, err := s.conversationRepo.GetConversationUsers(ctx, groupID)
	if err != nil {
		return err
	}

	err = s.groupRepo.DeleteGroup(ctx, groupID)
	if err != nil {
		return err
	}

	s.sender.Broadcast(SendItem{
		RequestID: uuid.New(),
		Event:     model.ConversationDeleteEvent,
		Content:   groupID,
		Recievers: members,
	})

	return nil
}

func (s *GroupService) RotateInviteCode(ctx context.Context, userID, groupID uuid.UUID) (string, error) {
	pkg.Logger.Info().
		Str("user id", userID.String()).
		Str("group id", groupID.String()).
		Msg("rotate invite code")

	if userID == uuid.Nil || groupID == uuid.Nil {
		return "", apperr.InvalidInput("group service", "user and group id are required", nil)
	}

	if err := s.requireOwner(ctx, groupID, userID); err != nil {
		return "", err
	}

	var lastErr error
	for range inviteCodeAttempts {
		code, err := pkg.NewInviteCode()
		if err != nil {
			return "", apperr.Internal("group service", err)
		}

		err = s.groupRepo.UpdateInviteCode(ctx, groupID, code)
		if err == nil {
			return code, nil
		}

		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Type == apperr.ErrConflict {
			lastErr = err
			continue
		}

		return "", err
	}

	return "", apperr.Internal("group service", lastErr)
}

func (s *GroupService) Update(ctx context.Context, userID, groupID uuid.UUID, params UpdateGroupProfileParams) (*model.Group, error) {
	pkg.Logger.Info().
		Str("user id", userID.String()).
		Str("group id", groupID.String()).
		Msg("updating group")

	if userID == uuid.Nil || groupID == uuid.Nil {
		return nil, apperr.InvalidInput("group service", "user and group id are required", nil)
	}

	if err := s.requireOwner(ctx, groupID, userID); err != nil {
		return nil, err
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if params.Name != nil {
		name := strings.TrimSpace(*params.Name)

		if name == "" {
			return nil, apperr.InvalidInput("group service", "group name is required", nil)
		}

		if len([]rune(name)) > maxGroupNameLength {
			return nil, apperr.InvalidInput("group service", "group name is too long", nil)
		}

		group.Name = name
	}

	if params.Bio != nil {
		bio := strings.TrimSpace(*params.Bio)

		if len([]rune(bio)) > maxGroupBioLength {
			return nil, apperr.InvalidInput("group service", "group bio is too long", nil)
		}

		group.Bio = bio
	}

	updated, err := s.groupRepo.Update(ctx, UpdateGroupParams{
		GroupID:   groupID,
		Name:      group.Name,
		Bio:       group.Bio,
		AvatarKey: group.AvatarKey,
	})
	if err != nil {
		return nil, err
	}

	updated.AvatarKey = s.fixUrl(ctx, updated.AvatarKey)
	s.broadcastGroupUpdate(ctx, updated)

	return updated, nil
}

func (s *GroupService) UpdateAvatar(ctx context.Context, userID, groupID uuid.UUID, file model.UploadedFile) (*model.Group, error) {
	pkg.Logger.Info().
		Str("user id", userID.String()).
		Str("group id", groupID.String()).
		Msg("updating group avatar")

	if userID == uuid.Nil || groupID == uuid.Nil {
		return nil, apperr.InvalidInput("group service", "user and group id are required", nil)
	}

	if err := s.requireOwner(ctx, groupID, userID); err != nil {
		return nil, err
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	media, err := s.storage.UploadAvatar(ctx, groupID, "group", file)
	if err != nil {
		return nil, err
	}

	updated, err := s.groupRepo.Update(ctx, UpdateGroupParams{
		GroupID:   groupID,
		Name:      group.Name,
		Bio:       group.Bio,
		AvatarKey: media.Key,
	})
	if err != nil {
		if deleteErr := s.storage.Delete(ctx, media.Key); deleteErr != nil {
			pkg.Logger.Warn().
				Str("key", media.Key).
				AnErr("error", deleteErr).
				Msg("failed to clean up uploaded group avatar")
		}

		return nil, err
	}

	if group.AvatarKey != "" {
		if err := s.storage.Delete(ctx, group.AvatarKey); err != nil {
			pkg.Logger.Warn().
				Str("key", group.AvatarKey).
				AnErr("error", err).
				Msg("failed to delete previous group avatar")
		}
	}

	updated.AvatarKey = s.fixUrl(ctx, updated.AvatarKey)
	s.broadcastGroupUpdate(ctx, updated)

	return updated, nil
}

func (s *GroupService) broadcastGroupUpdate(ctx context.Context, group *model.Group) {
	members, err := s.conversationRepo.GetConversationUsers(ctx, group.ID)
	if err != nil {
		pkg.Logger.Warn().
			Str("group id", group.ID.String()).
			AnErr("error", err).
			Msg("failed to load group members for update broadcast")
		return
	}

	s.sender.Broadcast(SendItem{
		RequestID: uuid.New(),
		Event:     model.GroupUpdateEvent,
		Content:   group,
		Recievers: members,
	})
}

func (s *GroupService) requireOwner(ctx context.Context, groupID, userID uuid.UUID) error {
	role, err := s.groupRepo.GetMemberRole(ctx, groupID, userID)
	if err != nil {
		return err
	}

	if role == "" {
		return apperr.InvalidInput("group service", "invalid group", nil)
	}

	if role != model.GroupMemberRoleOwner {
		return apperr.InvalidInput("group service", "you are not an owner", nil)
	}

	return nil
}

func (s *GroupService) fixUrl(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}

	url, err := s.storage.URL(ctx, key)
	if err != nil {
		return ""
	}

	return url
}
