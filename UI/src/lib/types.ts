export type User = {
  id: string
  username: string
  display_name: string
  avatar_url: string
  bio: string
  created_at: string
  updated_at: string
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

export type MessageType = 'text'

export type MessageText = {
  content: string
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
  last_message_text: string
  last_message_created_at: string
  updated_at: string
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
  created_at: string
  updated_at: string
  request_id?: string
  edited?: boolean
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

export type WSEvent =
  | { type: 'message.create'; request_id: string; data: MessageCreated }
  | { type: 'message.edit'; request_id: string; data: MessageEdited }
  | { type: 'message.delete'; request_id: string; data: MessageDeleted }

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
