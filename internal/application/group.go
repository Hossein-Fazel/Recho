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
	GroupRepo        GroupRepo
	ConversationRepo ConversationRepo
}

func NewGroupService(groupRepo GroupRepo, conversationRepo ConversationRepo) *GroupService {
	pkg.Logger.Info().Msg("Initializing Group service")

	return &GroupService{
		GroupRepo:        groupRepo,
		ConversationRepo: conversationRepo,
	}
}

type CreateGroupParams struct {
	Name       string
	Bio        string
	InviteCode string
	CreatedBy  uuid.UUID
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

		group, err := s.GroupRepo.Create(ctx, CreateGroupParams{
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

// PreviewByInviteCode returns the public details of the group behind an invite
// code, along with whether the requesting user is already a member.
func (s *GroupService) PreviewByInviteCode(ctx context.Context, userID uuid.UUID, code string) (*model.GroupPreview, error) {
	code = strings.TrimSpace(code)

	if code == "" {
		return nil, apperr.InvalidInput("group service", "invite code is required", nil)
	}

	preview, err := s.GroupRepo.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if userID != uuid.Nil {
		role, err := s.GroupRepo.GetMemberRole(ctx, preview.GroupID, userID)
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

	if err := s.GroupRepo.AddMember(ctx, preview.GroupID, userID); err != nil {
		return nil, err
	}

	preview.IsMember = true
	preview.MemberCount++

	return preview, nil
}
