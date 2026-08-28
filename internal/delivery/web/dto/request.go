package dto

import "github.com/google/uuid"

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type GetOrCreateDirectConversationRequest struct {
	UserID uuid.UUID `json:"user_id"`
}
