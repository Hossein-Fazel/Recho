package model

import (
	"time"

	"github.com/google/uuid"
)

type UserConversation struct {
	ConversationID       uuid.UUID
	ConversationType     string
	UserID               uuid.UUID
	Username             string
	DisplayName          string
	AvatarUrl            string
	GroupName            string
	GroupAvatarUrl       string
	LastMessageID        int64
	LastMessageContent   string
	LastMessageCreatedAt time.Time
	UpdatedAt            time.Time
}

type GroupMemberRole string

const (
	GroupMemberRoleOwner  GroupMemberRole = "owner"
	GroupMemberRoleAdmin  GroupMemberRole = "admin"
	GroupMemberRoleMember GroupMemberRole = "member"
)

type GroupMember struct {
	UserID      uuid.UUID
	Username    string
	DisplayName string
	AvatarURL   string
	Role        GroupMemberRole
}

type ConversationInfo struct {
	ID               uuid.UUID
	ConversationType string
	User             *UserInfo
	Group            *GroupInfo
}

type UserInfo struct {
	UserID      uuid.UUID
	Username    string
	DisplayName string
	AvatarUrl   string
	Bio         string
}

type GroupInfo struct {
	GroupID        uuid.UUID
	GroupName      string
	GroupAvatarUrl string
	GroupBio       string
	InviteCode     string
}
