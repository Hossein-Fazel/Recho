export type User = {
  id: string
  username: string
  display_name: string
  avatar_url: string
  bio: string
  created_at: string
  updated_at: string
  last_seen: string
}

export type AuthResponse = {
  message: string
  user: User
}

export type ErrResponse = {
  error: string
  type?: string
  module?: string
}

export type MessageType = 'text' | 'file'

/** File categories accepted by the backend media upload endpoint. */
export type FileCategory = 'image' | 'video' | 'voice' | 'file'

export type MessageText = {
  content: string
}

/**
 * File payload attached to a `file` message. Field names match the backend
 * JSON exactly (snake_case) — see `dto.FileMessage` / `dto.MessageMediaUploadResponse`.
 */
export type MessageFile = {
  key: string
  url?: string
  category: FileCategory
  content_type: string
  size: number
  file_name?: string
  caption?: string
}

export type Conversation = {
  conversation_id: string
  conversation_type: string
  user_id: string
  username: string
  display_name: string
  avatar_url: string
  group_name: string
  group_avatar_url: string
  last_message_id: number
  last_message_type?: MessageType
  /**
   * Category of the latest message when it is a file, sent by the backend
   * (`image`/`video`/`voice`/`file`). Empty or absent when the latest message
   * is not a file.
   */
  last_file_category?: FileCategory
  last_message_text: string
  last_message_created_at: string
  updated_at: string
  last_seen: string
}

export type UserConversationsResponse = {
  conversations?: Conversation[] | null
  Conversations?: Conversation[] | null
  next_cursor?: string
}

export type UserSearch = {
  id: string
  username: string
  display_name: string
  avatar_url: string
}

export type Message = {
  id: number
  conversation_id: string
  sender_id: string
  type: MessageType
  text?: MessageText | null
  file?: MessageFile | null
  created_at: string
  updated_at: string
  request_id?: string
  edited?: boolean
  /** Local-only fields used by optimistic file uploads. */
  upload_status?: 'uploading' | 'sending'
  upload_progress?: number
  local_preview_url?: string
}

export type GetConversationMessagesResponse = {
  messages: Message[] | null
  next_cursor: string
}

export type WSResponse = {
  type: string
  request_id: string
  data: unknown
}

export type MessageCreated = Message

export type MessageEdited = Message & {
  request_id?: string
}

export type MessageDeleted = {
  id: number
  conversation_id: string
  request_id?: string
}

export type ConversationDeleted = {
  conversation_id: string
}

export type GroupUpdated = {
  group_id: string
  name: string
  avatar_url: string
  bio: string
}

export type PresenceStatus = {
  user_id: string
  online: boolean
}

export type PresenceResponse = {
  statuses: PresenceStatus[]
}

export type PresenceUpdated = {
  user_id: string
  online: boolean
  last_seen?: string
}

export type WSEvent =
  | { type: 'message.create'; request_id: string; data: MessageCreated }
  | { type: 'message.edit'; request_id: string; data: MessageEdited }
  | { type: 'message.delete'; request_id: string; data: MessageDeleted }
  | { type: 'conversation.delete'; request_id: string; data: ConversationDeleted }
  | { type: 'group.update'; request_id: string; data: GroupUpdated }
  | { type: 'user.online'; request_id: string; data: PresenceUpdated }
  | { type: 'user.offline'; request_id: string; data: PresenceUpdated }

export type ConversationInfo = {
  conversation_id: string
  conversation_type: 'direct' | 'group'
  user?: UserInfo | null
  group?: GroupInfo | null
}

export type UserInfo = {
  id: string
  username: string
  display_name: string
  avatar_url: string
  bio: string
}

export type GroupInfo = {
  id: string
  name: string
  avatar_url: string
  bio: string
  /** Only sent to members allowed to share it (owner and admins). */
  invite_code?: string
  role?: GroupMemberRole
  member_count?: number
}

export type CreateGroupResponse = {
  conversation_id: string
  name: string
  bio: string
  avatar_url: string
  invite_code: string
  created_at: string
  updated_at: string
}

export type RotateInviteCodeResponse = {
  new_invite_code: string
}

export type UpdateGroupResponse = {
  id: string
  name: string
  avatar_url: string
  bio: string
  updated_at: string
}

export type UpdateGroupInput = {
  name?: string
  bio?: string
}

export type GroupPreview = {
  conversation_id: string
  name: string
  avatar_url: string
  bio: string
  member_count: number
  is_member: boolean
}

export type GroupMemberRole = 'owner' | 'admin' | 'member'

export type GroupMember = {
  user_id: string
  username: string
  display_name: string
  avatar_url: string
  role: GroupMemberRole
}

export type GroupMembersResponse = {
  members: GroupMember[] | null
  next_cursor?: string
}
