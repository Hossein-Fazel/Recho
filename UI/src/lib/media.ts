import type { FileCategory } from './types'

/**
 * Client-side mirror of the backend media policy
 * (`internal/application/storage_policy.go`). The backend stays authoritative;
 * these checks only spare the user a round-trip and give a clear message before
 * an upload starts. Keep in sync with the Go rules.
 */

const KB = 1024
const MB = 1024 * KB
const GB = 1024 * MB

export const IMAGE_MAX_BYTES = 25 * MB
export const VOICE_MAX_BYTES = 100 * MB
export const VIDEO_MAX_BYTES = 5 * GB
export const FILE_MAX_BYTES = 50 * GB

export const IMAGE_CONTENT_TYPES = [
  'image/jpeg',
  'image/png',
  'image/webp',
  'image/gif',
  'image/avif',
  'image/heic',
  'image/heif',
]

export const VOICE_CONTENT_TYPES = [
  'audio/mpeg',
  'audio/aac',
  'audio/mp4',
  'audio/ogg',
  'audio/opus',
  'audio/wav',
  'audio/webm',
  'audio/flac',
  'audio/amr',
]

export const VIDEO_CONTENT_TYPES = [
  'video/mp4',
  'video/webm',
  'video/quicktime',
  'video/3gpp',
  'video/3gpp2',
  'video/x-matroska',
]

type MediaRule = {
  maxBytes: number
  contentTypes: readonly string[] | null
  label: string
}

const rules: Record<FileCategory, MediaRule> = {
  image: {
    maxBytes: IMAGE_MAX_BYTES,
    contentTypes: IMAGE_CONTENT_TYPES,
    label: 'Image',
  },
  video: {
    maxBytes: VIDEO_MAX_BYTES,
    contentTypes: VIDEO_CONTENT_TYPES,
    label: 'Video',
  },
  voice: {
    maxBytes: VOICE_MAX_BYTES,
    contentTypes: VOICE_CONTENT_TYPES,
    label: 'Audio',
  },
  file: {
    maxBytes: FILE_MAX_BYTES,
    contentTypes: null,
    label: 'File',
  },
}

/**
 * Picks the upload category for a file. A file only counts as image/video/audio
 * when its browser-reported MIME type is one the backend accepts for that
 * category; anything else (including unsupported media types such as SVG or
 * BMP) is sent as a generic document so it still has a way through.
 */
export function categorizeFile(file: File): FileCategory {
  const type = (file.type || '').toLowerCase()

  if (type.startsWith('image/') && IMAGE_CONTENT_TYPES.includes(type)) {
    return 'image'
  }
  if (type.startsWith('video/') && VIDEO_CONTENT_TYPES.includes(type)) {
    return 'video'
  }
  if (type.startsWith('audio/') && VOICE_CONTENT_TYPES.includes(type)) {
    return 'voice'
  }

  return 'file'
}

export type FileValidation =
  | { ok: true; category: FileCategory }
  | { ok: false; category: FileCategory; error: string }

export function validateFile(file: File): FileValidation {
  const category = categorizeFile(file)
  const rule = rules[category]

  if (file.size <= 0) {
    return { ok: false, category, error: 'That file is empty' }
  }

  if (file.size > rule.maxBytes) {
    return {
      ok: false,
      category,
      error: `${rule.label} must be smaller than ${formatLimit(rule.maxBytes)}`,
    }
  }

  return { ok: true, category }
}

function formatLimit(bytes: number): string {
  if (bytes >= GB) return `${Math.round(bytes / GB)} GB`
  return `${Math.round(bytes / MB)} MB`
}

/** Whether the browser can safely render this file inline as a preview. */
export function isPreviewableCategory(category: FileCategory): boolean {
  return category === 'image' || category === 'video' || category === 'voice'
}
