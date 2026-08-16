package pkg

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	reset  = "\033[0m"
	white  = "\033[37m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	blue   = "\033[34m"
)

var Logger *slog.Logger

type PrettyHandler struct {
	level slog.Level
}

func NewPrettyHandler(level slog.Level) slog.Handler {
	return &PrettyHandler{level: level}
}

func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	levelText, levelColor := formatLevel(r.Level)

	fmt.Printf(
		"%s%s%s %s%s%s %s%s%s",
		white,
		r.Time.Format("15:04:05"),
		reset,

		levelColor,
		levelText,
		reset,

		"", r.Message, "",
	)

	r.Attrs(func(a slog.Attr) bool {
		fmt.Printf(
			" %s%s=%v%s",
			white,
			a.Key,
			a.Value.Any(),
			reset,
		)
		return true
	})

	fmt.Println()

	return nil
}

func (h *PrettyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *PrettyHandler) WithGroup(_ string) slog.Handler {
	return h
}

func formatLevel(level slog.Level) (string, string) {
	switch {
	case level >= slog.LevelError:
		return "ERROR", red
	case level >= slog.LevelWarn:
		return "WARN ", yellow
	case level >= slog.LevelInfo:
		return "INFO ", green
	default:
		return "DEBUG", blue
	}
}

func InitLogger(level slog.Level) {
	Logger = slog.New(
		NewPrettyHandler(level),
	)
}

func init() {
    Logger = slog.New(
        NewPrettyHandler(slog.LevelInfo),
    )
}
