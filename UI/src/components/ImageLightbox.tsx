import { useEffect } from 'react'
import { formatFileSize } from '../lib/format'
import type { MessageFile } from '../lib/types'

type ImageLightboxProps = {
  file: MessageFile
  onClose: () => void
}

/**
 * Full-screen viewer for an image message. Reuses the app's modal overlay
 * styling; clicking the backdrop or pressing Escape closes it.
 */
export function ImageLightbox({ file, onClose }: ImageLightboxProps) {
  useEffect(() => {
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])

  const url = file.url?.trim()
  if (!url) return null

  const name = file.file_name?.trim() || 'Image'

  return (
    <div
      className="modal-overlay lightbox-overlay"
      role="dialog"
      aria-modal="true"
      aria-label={name}
      onClick={onClose}
    >
      <figure className="lightbox" onClick={(event) => event.stopPropagation()}>
        <img src={url} alt={name} onError={onClose} />

        <figcaption className="lightbox-caption">
          <span className="lightbox-name" title={name}>
            {name}
          </span>
          <span className="lightbox-size">{formatFileSize(file.size)}</span>
          <a
            className="lightbox-open"
            href={url}
            target="_blank"
            rel="noreferrer noopener"
          >
            Open original
          </a>
        </figcaption>

        <button
          type="button"
          className="lightbox-close"
          onClick={onClose}
          aria-label="Close image"
          title="Close"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M6 6l12 12M18 6 6 18" />
          </svg>
        </button>
      </figure>
    </div>
  )
}
