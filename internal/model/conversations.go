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
	AvatarKey            string
	LastSeen             time.Time
	GroupName            string
	GroupAvatarKey       string
	LastMessageID        int64
	LastMessageType      MessageType
	LastMessageText      string
	LastFileCategory     MediaCategory
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
	AvatarKey   string
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
	AvatarKey   string
	Bio         string
}

type GroupInfo struct {
	GroupID        uuid.UUID
	GroupName      string
	GroupAvatarKey string
	GroupBio       string
	InviteCode     string
	ViewerRole     GroupMemberRole
	MemberCount    int64
}

func (g *GroupInfo) CanShareInvite() bool {
	return g.ViewerRole == GroupMemberRoleOwner ||
		g.ViewerRole == GroupMemberRoleAdmin
}

type Group struct {
	ID         uuid.UUID
	Name       string
	AvatarKey  string
	Bio        string
	InviteCode string
	CreatedBy  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type GroupPreview struct {
	GroupID     uuid.UUID
	Name        string
	AvatarKey   string
	Bio         string
	MemberCount int64
	IsMember    bool
}
