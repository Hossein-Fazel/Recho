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

type UpdateProfileRequest struct {
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
}

type GetOrCreateDirectConversationRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type CreateGroupRequest struct {
	Name string `json:"name"`
	Bio  string `json:"bio"`
}

type UpdateGroupRequest struct {
	Name *string `json:"name"`
	Bio  *string `json:"bio"`
}

type JoinGroupRequest struct {
	InviteCode string `json:"invite_code"`
}

type SubscribePresenceRequest struct {
	UserIDs []uuid.UUID `json:"user_ids"`
}

type UnsubscribePresenceRequest struct {
	UserIDs []uuid.UUID `json:"user_ids"`
}
