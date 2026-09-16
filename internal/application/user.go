package application

import (
	"context"
	"strings"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

const (
	userModule = "user service"

	maxDisplayNameLength = 100
	maxBioLength         = 250
)

type UpdateProfileParams struct {
	DisplayName *string
	Bio         *string
}

type UserService struct {
	userRepo UserRepo
	storage  *StorageService
}

func NewUserService(userRepo UserRepo, storage *StorageService) *UserService {
	pkg.Logger.Info().Msg("Initializing User service")

	return &UserService{
		userRepo: userRepo,
		storage:  storage,
	}
}

func (u *UserService) Search(ctx context.Context, username string) ([]*model.UserSearch, error) {
	if strings.TrimSpace(username) == "" {
		return []*model.UserSearch{}, apperr.InvalidInput(userModule, "username is required", nil)
	}
	users, err := u.userRepo.Search(ctx, username)
	if err != nil {
		return nil, err
	}

	for idx, user := range users {
		user.AvatarKey = u.fixUrl(ctx, user.AvatarKey)
		users[idx] = user
	}

	return users, nil
}

func (u *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	if userID == uuid.Nil {
		return nil, apperr.InvalidInput(userModule, "user id is required", nil)
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.AvatarKey = u.fixUrl(ctx, user.AvatarKey)
	return user, nil
}

func (u *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, params UpdateProfileParams) (*model.User, error) {
	if userID == uuid.Nil {
		return nil, apperr.InvalidInput(userModule, "user id is required", nil)
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if params.DisplayName != nil {
		displayName := strings.TrimSpace(*params.DisplayName)

		if len([]rune(displayName)) > maxDisplayNameLength {
			return nil, apperr.InvalidInput(userModule, "display name is too long", nil)
		}

		user.DisplayName = displayName
	}

	if params.Bio != nil {
		bio := strings.TrimSpace(*params.Bio)

		if len([]rune(bio)) > maxBioLength {
			return nil, apperr.InvalidInput(userModule, "bio is too long", nil)
		}

		user.Bio = bio
	}

	pkg.Logger.Info().
		Str("id", userID.String()).
		Msg("Updating user profile")

	updated, err := u.userRepo.Update(ctx, UpdateUserParams{
		ID:          userID,
		DisplayName: user.DisplayName,
		AvatarKey:   user.AvatarKey,
		Bio:         user.Bio,
	})
	if err != nil {
		return nil, err
	}

	updated.AvatarKey = u.fixUrl(ctx, updated.AvatarKey)

	return updated, nil
}

func (u *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, file model.UploadedFile) (*model.User, error) {
	if userID == uuid.Nil {
		return nil, apperr.InvalidInput(userModule, "user id is required", nil)
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	pkg.Logger.Info().
		Str("id", userID.String()).
		Msg("Updating user avatar")

	media, err := u.storage.UploadAvatar(ctx, userID, "user", file)
	if err != nil {
		return nil, err
	}

	updated, err := u.userRepo.Update(ctx, UpdateUserParams{
		ID:          userID,
		DisplayName: user.DisplayName,
		AvatarKey:   media.Key,
		Bio:         user.Bio,
	})
	if err != nil {
		if deleteErr := u.storage.Delete(ctx, media.Key); deleteErr != nil {
			pkg.Logger.Warn().
				Str("key", media.Key).
				AnErr("error", deleteErr).
				Msg("failed to clean up uploaded avatar")
		}

		return nil, err
	}

	if user.AvatarKey != "" {
		if err := u.storage.Delete(ctx, user.AvatarKey); err != nil {
			pkg.Logger.Warn().
				Str("url", user.AvatarKey).
				AnErr("error", err).
				Msg("failed to delete previous avatar")
		}
	}

	updated.AvatarKey = u.fixUrl(ctx, updated.AvatarKey)

	return updated, nil
}

func (u *UserService) fixUrl(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}

	url, err := u.storage.URL(ctx, key)
	if err != nil {
		return ""
	}
	return url
}
