import type { FileCategory, Message } from './types'

export const UNKNOWN_USER = 'Unknown user'
export const UNNAMED_GROUP = 'Unnamed group'

/** Text content of a message, or '' for non-text/missing payloads. */
export function messageText(message: Message): string {
  return message.text?.content ?? ''
}

/**
 * Human-readable label for a file message's category, shown in the chat list
 * preview. Unknown or missing categories fall back to a generic document so a
 * row never leaks a filename, MIME type, size, key, or URL.
 */
export function fileCategoryLabel(category?: FileCategory): string {
  switch (category) {
    case 'image':
      return 'Photo'
    case 'video':
      return 'Video'
    case 'voice':
      return 'Audio'
    default:
      return 'Document'
  }
}

/**
 * Short preview for a conversation list / sidebar row. File messages show the
 * category (e.g. `Photo`) instead of the caption, filename, or raw payload.
 */
export function messagePreview(message: Message): string {
  if (message.type === 'file') {
    return fileCategoryLabel(message.file?.category)
  }

  return messageText(message)
}

/** Human-readable file size, e.g. `2.4 MB`. */
export function formatFileSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'

  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const exponent = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    units.length - 1,
  )
  const value = bytes / 1024 ** exponent

  // Keep whole bytes whole, one decimal for everything else below 10.
  const digits = exponent === 0 ? 0 : value >= 10 ? 0 : 1

  return `${value.toFixed(digits)} ${units[exponent]}`
}

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


export function formatLastSeen(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime()) || date.getFullYear() < 2) return 'recently'

  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour

  if (diff >= 0 && diff < minute) return 'just now'
  if (diff >= 0 && diff < hour) {
    const minutes = Math.max(1, Math.floor(diff / minute))
    return `${minutes} minute${minutes === 1 ? '' : 's'} ago`
  }
  if (diff >= 0 && diff < day) {
    const hours = Math.floor(diff / hour)
    return `${hours} hour${hours === 1 ? '' : 's'} ago`
  }

  const sameYear = date.getFullYear() === now.getFullYear()
  return date.toLocaleString([], {
    month: 'short',
    day: 'numeric',
    ...(sameYear ? {} : { year: 'numeric' }),
    hour: 'numeric',
    minute: '2-digit',
  })
}
