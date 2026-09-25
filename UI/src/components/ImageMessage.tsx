import { useState } from 'react'
import { DocumentMessage } from './DocumentMessage'
import type { MessageFile } from '../lib/types'

type ImageMessageProps = {
  file: MessageFile
  onOpen: (file: MessageFile) => void
}

/**
 * Inline image preview. Clicking opens the lightbox. A missing or broken URL
 * degrades to the document card so the message never collapses to nothing.
 */
export function ImageMessage({ file, onOpen }: ImageMessageProps) {
  const [failed, setFailed] = useState(false)

  const url = file.url?.trim()
  if (!url || failed) {
    return <DocumentMessage file={file} />
  }

  const name = file.file_name?.trim() || 'Image'

  return (
    <button
      type="button"
      className="file-image"
      onClick={() => onOpen(file)}
      aria-label={`Open image ${name}`}
      title="Open image"
    >
      <img
        src={url}
        alt={name}
        loading="lazy"
        onError={() => setFailed(true)}
      />
    </button>
  )
}
