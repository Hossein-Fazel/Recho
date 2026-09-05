export const UNKNOWN_USER = 'Unknown user'
export const UNNAMED_GROUP = 'Unnamed group'

type NameFields = {
  display_name?: string
  username?: string
  group_name?: string
  name?: string
}

/**
 * Best available name, preferring a real name and falling back to the
 * username. Returns '' when nothing is set.
 */
export function displayName(input: NameFields): string {
  return (
    input.display_name?.trim() ||
    input.group_name?.trim() ||
    input.name?.trim() ||
    input.username?.trim() ||
    ''
  )
}

/** Name of a person, never empty. Falls back to the username. */
export function personName(input: NameFields): string {
  return displayName(input) || UNKNOWN_USER
}

/** Title of a direct or group conversation, never empty. */
export function conversationTitle(
  input: NameFields & { conversation_type?: string },
): string {
  return (
    displayName(input) ||
    (input.conversation_type === 'group' ? UNNAMED_GROUP : UNKNOWN_USER)
  )
}

/**
 * `@username` for use as a secondary line. Empty when there is no username,
 * or when `name` is already the username and would just be repeated.
 */
export function handle(username?: string, name?: string): string {
  const value = username?.trim()
  if (!value) return ''
  if (name?.trim() === value) return ''
  return `@${value}`
}

export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return '?'
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[1][0]).toUpperCase()
}

export function hueFromId(id: string): number {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = (hash * 31 + id.charCodeAt(i)) >>> 0
  }
  return hash % 360
}

export function formatTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2) return ''

  const now = new Date()
  const sameDay =
    date.getDate() === now.getDate() &&
    date.getMonth() === now.getMonth() &&
    date.getFullYear() === now.getFullYear()

  if (sameDay) {
    return date.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })
  }

  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (
    date.getDate() === yesterday.getDate() &&
    date.getMonth() === yesterday.getMonth() &&
    date.getFullYear() === yesterday.getFullYear()
  ) {
    return 'Yesterday'
  }

  return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

export function formatMessageTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2) {
    return ''
  }
  return date.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })
}

/** Full date and time. Returns '' for missing or zero-value timestamps. */
export function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2) {
    return ''
  }
  return date.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}
