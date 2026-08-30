package pkg

import (
	"os"

	"github.com/rs/zerolog"
)

var Logger *zerolog.Logger

func init() {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}
	level := zerolog.InfoLevel

	logger := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Logger()

	Logger = &logger
}
