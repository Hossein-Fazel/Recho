package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/pkg"
)

const defaultDirectoryPermision = 0750 // Owner: RWX, Group: R-X, Other: ---

type LocalConfig struct {
	RootPath string `env:"ROOT_PATH"`
	BaseURL  string `env:"BASE_URL"`
}

type Local struct {
	cfg LocalConfig
}

var _ application.Storage = (*Local)(nil)

func NewLocal(cfg LocalConfig) (application.Storage, error) {
	if err := os.MkdirAll(cfg.RootPath, defaultDirectoryPermision); err != nil {
		return nil, apperr.InvalidInput("storage", "failed to create local storage root path", err)
	}

	return &Local{
		cfg: cfg,
	}, nil
}

func (a *Local) Upload(_ context.Context, file io.Reader, _ int64, key, _ string, _ bool) (string, error) {
	fullPath := filepath.Join(a.cfg.RootPath, key)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, defaultDirectoryPermision); err != nil {
		return "", apperr.Internal("storage", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", apperr.Internal("storage", err)
	}
	defer func() {
		if err := dst.Close(); err != nil {
			pkg.Logger.Debug().Msg("failed to close")
		}
	}()

	if _, err := io.Copy(dst, file); err != nil {
		return "", apperr.Internal("storage", err)
	}

	return key, nil
}

func (a *Local) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(a.cfg.RootPath, key)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return apperr.Internal("storage", err)
	}

	return nil
}

func (a *Local) GetURL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("%s/%s", a.cfg.BaseURL, key), nil
}

func (a *Local) GetPresignedURL(ctx context.Context, key string) (string, error) {
	return a.GetURL(ctx, key)
}

func (a *Local) Exists(_ context.Context, key string) (bool, error) {
	fullPath := filepath.Join(a.cfg.RootPath, key)

	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
