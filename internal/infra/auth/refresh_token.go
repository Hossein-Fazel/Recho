package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

)

type RefreshTokenService struct{}

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{}
}

func (s *RefreshTokenService) Generate() (plain string, hash string, err error) {
	raw := make([]byte, 32)

	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}

	plain = base64.RawURLEncoding.EncodeToString(raw)

	sum := sha256.Sum256(raw)

	hash = base64.RawURLEncoding.EncodeToString(sum[:])

	return plain, hash, nil
}