package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

const (
	storageModule = "storage service"

	publicPrefix       = "avatars"
	userAvatarPrefix   = publicPrefix + "/users"
	groupAvatarPrefix  = publicPrefix + "/groups"
	conversationPrefix = "conversations"

	maxFileNameLength = 255
)

type UploadFile struct {
	Content  io.Reader
	Size     int64
	FileName string
}

type StorageService struct {
	storage application.Storage
}

func NewStorageService(storage application.Storage) *StorageService {
	pkg.Logger.Info().Msg("Initializing Storage service")

	return &StorageService{
		storage: storage,
	}
}

func (s *StorageService) UploadAvatar(ctx context.Context, ID uuid.UUID, file UploadFile) (*model.Media, error) {
	if ID == uuid.Nil {
		return nil, apperr.InvalidInput(storageModule, "id is required", nil)
	}

	pkg.Logger.Info().
		Str("id", ID.String()).
		Msg("Uploading avatar")

	return s.upload(ctx, path.Join(userAvatarPrefix, ID.String()), model.MediaCategoryAvatar, file)
}

func (s *StorageService) UploadMessageMedia(
	ctx context.Context,
	conversationID uuid.UUID,
	category model.MediaCategory,
	file UploadFile,
) (*model.Media, error) {
	if conversationID == uuid.Nil {
		return nil, apperr.InvalidInput(storageModule, "conversation id is required", nil)
	}

	if category == model.MediaCategoryAvatar {
		return nil, apperr.InvalidInput(storageModule, "avatars cannot be attached to a conversation", nil)
	}

	pkg.Logger.Info().
		Str("conversation id", conversationID.String()).
		Str("category", category.String()).
		Msg("Uploading message media")

	now := time.Now().UTC()
	prefix := path.Join(
		conversationPrefix,
		conversationID.String(),
		category.String(),
		now.Format("2006"),
		now.Format("01"),
	)

	return s.upload(ctx, prefix, category, file)
}

func (s *StorageService) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	return s.storage.Delete(ctx, key)
}

func (s *StorageService) DeleteAll(ctx context.Context, keys ...string) error {
	var firstErr error

	for _, key := range keys {
		if err := s.Delete(ctx, key); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (s *StorageService) URL(ctx context.Context, key string) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}

	return s.urlFor(ctx, key, isPublicKey(key))
}

func (s *StorageService) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}

	return s.storage.Exists(ctx, key)
}

func (s *StorageService) upload(
	ctx context.Context,
	prefix string,
	category model.MediaCategory,
	file UploadFile,
) (*model.Media, error) {
	rule, ok := mediaRules[category]
	if !ok {
		return nil, apperr.InvalidInput(storageModule, "unsupported media category: "+category.String(), nil)
	}

	if file.Content == nil {
		return nil, apperr.InvalidInput(storageModule, "file content is required", nil)
	}

	if file.Size > rule.MaxSize {
		return nil, apperr.InvalidInput(storageModule, sizeLimitMessage(category, rule.MaxSize), nil)
	}

	header, err := readHeader(file.Content)
	if err != nil {
		return nil, err
	}

	contentType := canonicalContentType(category, pkg.DetectContentType(header))
	if !rule.Allows(contentType) {
		return nil, apperr.InvalidInput(
			storageModule,
			fmt.Sprintf("unsupported %s content type: %s", category, contentType),
			nil,
		)
	}

	key := path.Join(prefix, uuid.NewString()+pkg.ExtensionByContentType(contentType))

	body := &model.LimitedReader{
		Reader: io.MultiReader(bytes.NewReader(header), file.Content),
		Max:    rule.MaxSize,
	}

	if _, err := s.storage.Upload(ctx, body, file.Size, key, contentType, rule.Public); err != nil {
		s.discard(ctx, key)

		if errors.Is(err, model.ErrFileTooLarge) {
			return nil, apperr.InvalidInput(storageModule, sizeLimitMessage(category, rule.MaxSize), nil)
		}

		return nil, err
	}

	url, err := s.urlFor(ctx, key, rule.Public)
	if err != nil {
		s.discard(ctx, key)

		return nil, err
	}

	return &model.Media{
		Key:         key,
		URL:         url,
		Category:    category,
		ContentType: contentType,
		Size:        body.Readed,
		FileName:    sanitizeFileName(file.FileName),
		Public:      rule.Public,
		UploadedAt:  time.Now().UTC(),
	}, nil
}

func (s *StorageService) urlFor(ctx context.Context, key string, public bool) (string, error) {
	if public {
		return s.storage.GetURL(ctx, key)
	}

	return s.storage.GetPresignedURL(ctx, key)
}

func (s *StorageService) discard(ctx context.Context, key string) {
	if err := s.storage.Delete(ctx, key); err != nil {
		pkg.Logger.Warn().
			Str("key", key).
			AnErr("error", err).
			Msg("failed to clean up incomplete upload")
	}
}

func validateKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return apperr.InvalidInput(storageModule, "object key is required", nil)
	}

	if key != strings.TrimSpace(key) || strings.HasPrefix(key, "/") || strings.Contains(key, `\`) {
		return apperr.InvalidInput(storageModule, "invalid object key", nil)
	}

	for _, r := range key {
		if r < 0x20 || r == 0x7F {
			return apperr.InvalidInput(storageModule, "invalid object key", nil)
		}
	}

	for _, element := range strings.Split(key, "/") {
		if element == "" || element == "." || element == ".." {
			return apperr.InvalidInput(storageModule, "invalid object key", nil)
		}
	}

	return nil
}

func canonicalContentType(category model.MediaCategory, contentType string) string {
	if category != model.MediaCategoryVoice {
		return contentType
	}

	switch contentType {
	case "video/webm":
		return "audio/webm"
	case "video/mp4":
		return "audio/mp4"
	case "video/ogg", "application/ogg":
		return "audio/ogg"
	default:
		return contentType
	}
}

func isPublicKey(key string) bool {
	return strings.HasPrefix(key, publicPrefix+"/")
}

func sizeLimitMessage(category model.MediaCategory, maxSize int64) string {
	return fmt.Sprintf("%s must not exceed %d MB", category, maxSize>>20)
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	// Drop any directory component a client may have sent.
	name = name[strings.LastIndexAny(name, `/\`)+1:]

	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return -1
		}

		return r
	}, name)

	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return ""
	}

	if runes := []rune(name); len(runes) > maxFileNameLength {
		name = string(runes[:maxFileNameLength])
	}

	return name
}

func readHeader(content io.Reader) ([]byte, error) {
	header := make([]byte, 512)

	n, err := io.ReadFull(content, header)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, apperr.InvalidInput("storage service", "failed to read file content", err)
	}

	if n == 0 {
		return nil, apperr.InvalidInput("storage service", "file is empty", nil)
	}

	return header[:n], nil
}
