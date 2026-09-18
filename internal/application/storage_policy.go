package application

import "github.com/Hossein-Fazel/Recho/internal/model"

const (
	maxAvatarSize = 5 << 20   // 5 MB
	maxImageSize  = 25 << 20  // 25 MB
	maxVoiceSize  = 100 << 20 // 100 MB
	maxVideoSize  = 5 << 30   // 5 GB
	maxFileSize   = 50 << 30  // 50 GB
)

type mediaRule struct {
	MaxSize      int64
	Public       bool
	AllowedTypes map[string]struct{}
}

func (r mediaRule) Allows(contentType string) bool {
	if r.AllowedTypes == nil {
		return true
	}

	_, ok := r.AllowedTypes[contentType]
	return ok
}

var mediaRules = map[model.MediaCategory]mediaRule{
	model.MediaCategoryAvatar: {
		MaxSize: maxAvatarSize,
		Public:  true,
		AllowedTypes: contentTypeSet(
			"image/jpeg",
			"image/png",
			"image/webp",
			"image/gif",
		),
	},

	model.MediaCategoryImage: {
		MaxSize: maxImageSize,
		AllowedTypes: contentTypeSet(
			"image/jpeg",
			"image/png",
			"image/webp",
			"image/gif",
			"image/avif",
			"image/heic",
			"image/heif",
		),
	},

	model.MediaCategoryVoice: {
		MaxSize: maxVoiceSize,
		AllowedTypes: contentTypeSet(
			"audio/mpeg",
			"audio/aac",
			"audio/mp4",
			"audio/ogg",
			"audio/opus",
			"audio/wav",
			"audio/webm",
			"audio/flac",
			"audio/amr",
		),
	},

	model.MediaCategoryVideo: {
		MaxSize: maxVideoSize,
		AllowedTypes: contentTypeSet(
			"video/mp4",
			"video/webm",
			"video/quicktime",
			"video/3gpp",
			"video/3gpp2",
			"video/x-matroska",
		),
	},

	model.MediaCategoryFile: {
		MaxSize: maxFileSize,
	},
}

func contentTypeSet(types ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(types))

	for _, typ := range types {
		result[typ] = struct{}{}
	}

	return result
}
