import { firstConvCursor, firstGroupMemberCursor, firstMessageCursor } from './cursor'
import type {
  AuthResponse,
  Conversation,
  ConversationInfo,
  CreateGroupResponse,
  ErrResponse,
  GetConversationMessagesResponse,
  GroupMember,
  GroupMembersResponse,
  GroupPreview,
  Message,
  RotateInviteCodeResponse,
  UpdateGroupInput,
  UpdateGroupResponse,
  User,
  UserConversationsResponse,
  UserSearch,
} from './types'

export type UpdateProfileInput = {
  username?: string
  display_name?: string
  bio?: string
}

class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

async function parseError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as ErrResponse
    return body.error || res.statusText
  } catch {
    return res.statusText
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  // Let the browser set the multipart boundary for FormData bodies; forcing
  // JSON there would make the server reject the request.
  const isFormData = init?.body instanceof FormData

  const res = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
      ...(init?.headers ?? {}),
    },
  })

  if (!res.ok) {
    throw new ApiError(await parseError(res), res.status)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return (await res.json()) as T
}

export const api = {
  login(username: string, password: string) {
    return request<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
  },

  register(username: string, password: string) {
    return request<AuthResponse>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
  },

  logout() {
    return request<{ message: string }>('/api/auth/logout', { method: 'POST' })
  },

  verify() {
    return request<{ message: string }>('/api/auth/verify')
  },

  refresh() {
    return request<{ message: string }>('/api/auth/refresh')
  },

  async conversations(cursor = firstConvCursor, limit = 30) {
    const params = new URLSearchParams({
      limit: String(limit),
      cursor,
    })
    const data = await request<UserConversationsResponse>(
      `/api/conversation/?${params}`,
    )
    const list = data.conversations ?? data.Conversations ?? []
    return {
      conversations: list.filter((c): c is Conversation => Boolean(c)),
      nextCursor: data.next_cursor ?? '',
    }
  },

  searchUsers(q: string) {
    const params = new URLSearchParams({ q })
    return request<{ users: UserSearch[] | null }>(`/api/user/search?${params}`)
  },

  getMe() {
    return request<User>('/api/user/me')
  },

  updateProfile(patch: UpdateProfileInput) {
    return request<User>('/api/user/me', {
      method: 'PATCH',
      body: JSON.stringify(patch),
    })
  },

  updateAvatar(file: File) {
    const body = new FormData()
    body.append('avatar', file)
    return request<User>('/api/user/me/avatar', {
      method: 'POST',
      body,
    })
  },

  conversation(conversationId: string) {
    return request<Conversation>(`/api/conversation/${conversationId}`)
  },

  openDirect(userId: string) {
    return request<{ conversation_id: string }>('/api/conversation/', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    })
  },

  async messages(conversationId: string, cursor = firstMessageCursor, limit = 30) {
    const params = new URLSearchParams({
      limit: String(limit),
      cursor_id: cursor,
    })
    const data = await request<GetConversationMessagesResponse>(
      `/api/conversation/${conversationId}/messages?${params}`,
    )
    return {
      messages: (data.messages ?? []).filter((m): m is Message => Boolean(m)),
      nextCursor: data.next_cursor ?? '',
    }
  },

  conversationInfo(conversationId: string) {
    return request<ConversationInfo>(
      `/api/conversation/${conversationId}/info`,
    )
  },

  async groupMembers(conversationId: string, cursor = firstGroupMemberCursor, limit = 50) {
    const params = new URLSearchParams({
      limit: String(limit),
      cursor,
    })
    const data = await request<GroupMembersResponse>(
      `/api/conversation/${conversationId}/members?${params}`,
    )
    return {
      members: (data.members ?? []).filter((m): m is GroupMember => Boolean(m)),
      nextCursor: data.next_cursor ?? '',
    }
  },

  createGroup(name: string, bio = '') {
    return request<CreateGroupResponse>('/api/group/', {
      method: 'POST',
      body: JSON.stringify({ name, bio }),
    })
  },

  groupByInviteCode(code: string) {
    return request<GroupPreview>(
      `/api/group/invite/${encodeURIComponent(code)}`,
    )
  },

  joinGroup(code: string) {
    return request<GroupPreview>('/api/group/join', {
      method: 'POST',
      body: JSON.stringify({ invite_code: code }),
    })
  },

  leaveGroup(groupId: string) {
    return request<{ message: string }>(
      `/api/group/${encodeURIComponent(groupId)}/leave`,
      { method: 'POST' },
    )
  },

  deleteGroup(groupId: string) {
    return request<{ message: string }>(
      `/api/group/${encodeURIComponent(groupId)}`,
      { method: 'DELETE' },
    )
  },

  rotateInviteCode(groupId: string) {
    return request<RotateInviteCodeResponse>(
      `/api/group/${encodeURIComponent(groupId)}/invite-code/rotate`,
      { method: 'POST' },
    )
  },

  updateGroup(groupId: string, patch: UpdateGroupInput) {
    return request<UpdateGroupResponse>(
      `/api/group/${encodeURIComponent(groupId)}`,
      {
        method: 'PATCH',
        body: JSON.stringify(patch),
      },
    )
  },

  updateGroupAvatar(groupId: string, file: File) {
    const body = new FormData()
    body.append('avatar', file)
    return request<UpdateGroupResponse>(
      `/api/group/${encodeURIComponent(groupId)}/avatar`,
      {
        method: 'POST',
        body,
      },
    )
  },
}

const USER_KEY = 'recho.user'

export function saveUser(user: User) {
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function loadUser(): User | null {
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? (JSON.parse(raw) as User) : null
  } catch {
    return null
  }
}

export function clearUser() {
  localStorage.removeItem(USER_KEY)
}

export { ApiError }
