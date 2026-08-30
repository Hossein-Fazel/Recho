package token

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

func (s *RefreshTokenService) Generate() (*model.RefreshToken, *model.UserRefreshToken, error) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return nil, nil, err
	}

	plain := base64.RawURLEncoding.EncodeToString(raw)

	hash := s.Hash(plain)

	expiresAt := time.Now().Add(s.ttl * 24 * time.Hour)

	return &model.RefreshToken{
		Hash:      hash,
		ExpiresAt: expiresAt,
	}, &model.UserRefreshToken{
		Token: plain,
		TTL:   s.ttl,
	}, nil
}

func (s *RefreshTokenService) Hash(plainText string) string {
	sum := sha256.Sum256([]byte(plainText))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
