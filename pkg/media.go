package pkg

import (
	"net/http"
	"strings"
)

const (
	DefaultContentType = "application/octet-stream"
	DefaultExtension   = ".bin"
)

var contentTypeExtensions = map[string]string{
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/webp":    ".webp",
	"image/gif":     ".gif",
	"image/avif":    ".avif",
	"image/heic":    ".heic",
	"image/heif":    ".heif",
	"image/bmp":     ".bmp",
	"image/tiff":    ".tiff",
	"image/svg+xml": ".svg",

	"audio/mpeg":  ".mp3",
	"audio/aac":   ".aac",
	"audio/ogg":   ".ogg",
	"audio/wav":   ".wav",
	"audio/x-wav": ".wav",
	"audio/mp4":   ".m4a",
	"audio/webm":  ".webm",
	"audio/opus":  ".opus",
	"audio/flac":  ".flac",
	"audio/amr":   ".amr",
	"audio/3gpp":  ".3gp",

	"video/mp4":        ".mp4",
	"video/webm":       ".webm",
	"video/quicktime":  ".mov",
	"video/x-matroska": ".mkv",
	"video/3gpp":       ".3gp",
	"video/3gpp2":      ".3g2",

	"application/pdf":              ".pdf",
	"application/zip":              ".zip",
	"application/gzip":             ".gz",
	"application/x-tar":            ".tar",
	"application/x-7z-compressed":  ".7z",
	"application/x-rar-compressed": ".rar",
	"application/msword":           ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"application/vnd.ms-excel": ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
	"application/vnd.ms-powerpoint":                                             ".ppt",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
	"application/rtf":  ".rtf",
	"application/json": ".json",
	"application/xml":  ".xml",
	"text/plain":       ".txt",
	"text/csv":         ".csv",
	"text/html":        ".html",
	"text/markdown":    ".md",
	"text/xml":         ".xml",
}

func DetectContentType(header []byte) string {
	if len(header) == 0 {
		return DefaultContentType
	}

	return http.DetectContentType(header)
}

func ExtensionByContentType(contentType string) string {
	mediaType := strings.ToLower(strings.TrimSpace(contentType))
	if idx := strings.IndexByte(mediaType, ';'); idx != -1 {
		mediaType = strings.TrimSpace(mediaType[:idx])
	}

	if ext, ok := contentTypeExtensions[mediaType]; ok {
		return ext
	}

	return DefaultExtension
}
