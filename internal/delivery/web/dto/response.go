package dto

import "github.com/Hossein-Fazel/Recho/internal/model"

type ErrResponse struct {
	Error  string `json:"error"`
	Type   string `json:"type,omitempty"`
	Module string `json:"module,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type AuthResponse struct {
	Message string      `json:"message"`
	User    *model.User `json:"user"`
}
