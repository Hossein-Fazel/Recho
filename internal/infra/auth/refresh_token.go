package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/model"
)

type RefreshTokenService struct {
	ttl time.Duration
}

func NewRefreshTokenService(conf Config) *RefreshTokenService {
	return &RefreshTokenService{
		ttl: conf.RefreshTTL,
	}
}

func (s *RefreshTokenService) Generate() (*model.RefreshToken, string, error) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return nil, "", err
	}

	plain := base64.RawURLEncoding.EncodeToString(raw)

	sum := sha256.Sum256([]byte(plain))
	hash := base64.RawURLEncoding.EncodeToString(sum[:])

	expiresAt := time.Now().Add(s.ttl * 24 * time.Hour)

	return &model.RefreshToken{
		Hash:      hash,
		ExpiresAt: expiresAt,
	}, plain, nil
}
