package pkg

import (
	"net/http"
)

const (
	DefaultContentType = "application/octet-stream"
	DefaultExtension   = ".bin"
)

var contentTypeExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",

	"audio/mpeg": ".mp3",
	"audio/ogg":  ".ogg",
	"audio/wav":  ".wav",
	"audio/mp4":  ".m4a",

	"video/mp4":  ".mp4",
	"video/webm": ".webm",
}

func DetectContentType(header []byte) string {
	if len(header) == 0 {
		return DefaultContentType
	}

	return http.DetectContentType(header)
}

func ExtensionByContentType(contentType string) string {
	if ext, ok := contentTypeExtensions[contentType]; ok {
		return ext
	}

	return DefaultExtension
}
