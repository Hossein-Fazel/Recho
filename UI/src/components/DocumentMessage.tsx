import { formatFileSize } from '../lib/format'
import type { MessageFile } from '../lib/types'

type DocumentMessageProps = {
  /** May be absent when a file message arrives without a payload. */
  file?: MessageFile
}

/**
 * Compact card for generic documents (and the fallback whenever a media URL is
 * missing or fails to load). Filenames are rendered as text, so React escapes
 * them — never inject the name as HTML.
 */
export function DocumentMessage({ file }: DocumentMessageProps) {
  const name = file?.file_name?.trim() || 'File'
  const size = formatFileSize(file?.size ?? 0)
  const url = file?.url?.trim()

  const body = (
    <>
      <span className="file-card-icon" aria-hidden>
        <svg viewBox="0 0 24 24">
          <path d="M6 3.5h7L18 8.5v12a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1v-16a1 1 0 0 1 1-1Z" />
          <path d="M13 3.5V8.5H18" />
        </svg>
      </span>

      <span className="file-card-meta">
        <span className="file-card-name" title={name}>
          {name}
        </span>
        <span className="file-card-size">{size}</span>
      </span>

      {url ? (
        <span className="file-card-action" aria-hidden>
          <svg viewBox="0 0 24 24">
            <path d="M12 4v11" />
            <path d="m8 11 4 4 4-4" />
            <path d="M5 19h14" />
          </svg>
        </span>
      ) : null}
    </>
  )

  if (!url) {
    return (
      <div className="file-card unavailable" title="This file is no longer available">
        {body}
      </div>
    )
  }

  return (
    <a
      className="file-card"
      href={url}
      target="_blank"
      rel="noreferrer noopener"
      title={`Open ${name}`}
    >
      {body}
    </a>
  )
}
