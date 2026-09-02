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
  last_message_content: string
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
  content: string
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
  | { type: 'error'; request_id: string; data: { code: number; message: string } }
