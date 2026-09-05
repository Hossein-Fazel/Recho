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
  User,
  UserConversationsResponse,
  UserSearch,
} from './types'

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
  const res = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
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
