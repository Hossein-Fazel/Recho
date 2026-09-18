package storage

import (
	"context"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
)

type Driver string

const (
	DriverLocal Driver = "local"
	DriverS3    Driver = "s3"
)

// Config selects and configures the storage backend.
type Config struct {
	Driver Driver      `env:"DRIVER" envDefault:"local"`
	Local  LocalConfig `envPrefix:"LOCAL_"`
	S3     S3Config    `envPrefix:"S3_"`
}

// New builds the Storage implementation selected by Config.Driver.
func New(ctx context.Context, cfg Config) (application.Storage, error) {
	switch cfg.Driver {
	case DriverLocal, "":
		return NewLocal(cfg.Local)
	case DriverS3:
		return NewS3(ctx, cfg.S3)
	default:
		return nil, apperr.InvalidInput("storage", "unsupported storage driver: "+string(cfg.Driver), nil)
	}
}
