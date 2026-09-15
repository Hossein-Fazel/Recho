package model

import (
	"errors"
	"io"
	"time"
)

type MediaCategory string

const (
	MediaCategoryAvatar MediaCategory = "avatar"
	MediaCategoryImage  MediaCategory = "image"
	MediaCategoryVoice  MediaCategory = "voice"
	MediaCategoryVideo  MediaCategory = "video"
	MediaCategoryFile   MediaCategory = "file"
)

var ErrFileTooLarge error = errors.New("file exceeds the maximum allowed size")

func (c MediaCategory) IsValid() bool {
	switch c {
	case MediaCategoryAvatar,
		MediaCategoryImage,
		MediaCategoryVoice,
		MediaCategoryVideo,
		MediaCategoryFile:
		return true
	default:
		return false
	}
}

func (c MediaCategory) String() string {
	return string(c)
}

type Media struct {
	Key         string
	URL         string
	Category    MediaCategory
	ContentType string
	Size        int64
	FileName    string
	Public      bool
	UploadedAt  time.Time
}

type LimitedReader struct {
	Reader io.Reader
	Max    int64
	Readed   int64
}

func (l *LimitedReader) Read(p []byte) (int, error) {
	remaining := l.Max - l.Readed

	if remaining <= 0 {
		return 0, ErrFileTooLarge
	}

	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err := l.Reader.Read(p)
	l.Readed += int64(n)

	return n, err
}
