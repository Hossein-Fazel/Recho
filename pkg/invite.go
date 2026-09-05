package pkg

import (
	"crypto/rand"
	"encoding/base64"
)

const InviteCodeLength = 12

func NewInviteCode() (string, error) {
	raw := make([]byte, InviteCodeLength)

	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}
