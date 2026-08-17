package pkg

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func Init() {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	level := zerolog.InfoLevel
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"

	logger := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	Logger = logger
}
