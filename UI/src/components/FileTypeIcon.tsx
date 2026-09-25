import type { FileCategory } from '../lib/types'

type FileTypeIconProps = {
  /** Backend file category, or undefined when it is not known client-side. */
  category?: FileCategory
}

/**
 * Small inline icon for a file message's category, used in the chat list
 * preview. Mirrors the SVG icon style used by the file message components and
 * falls back to the document glyph for unknown/missing categories.
 */
export function FileTypeIcon({ category }: FileTypeIconProps) {
  switch (category) {
    case 'image':
      return (
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M4 8.5A1.5 1.5 0 0 1 5.5 7h2L8.7 5h6.6L16.5 7h2A1.5 1.5 0 0 1 20 8.5v9a1.5 1.5 0 0 1-1.5 1.5h-13A1.5 1.5 0 0 1 4 17.5Z" />
          <circle cx="12" cy="12.5" r="3.2" />
        </svg>
      )
    case 'video':
      return (
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <rect x="3.5" y="5.5" width="17" height="13" rx="2.5" />
          <path d="m10.5 9.5 4.5 2.5-4.5 2.5Z" />
        </svg>
      )
    case 'voice':
      return (
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <rect x="9" y="3" width="6" height="11" rx="3" />
          <path d="M6 11a6 6 0 0 0 12 0" />
          <path d="M12 17v4M9 21h6" />
        </svg>
      )
    default:
      return (
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M6 3.5h7L18 8.5v12a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1v-16a1 1 0 0 1 1-1Z" />
          <path d="M13 3.5V8.5H18" />
        </svg>
      )
  }
}
