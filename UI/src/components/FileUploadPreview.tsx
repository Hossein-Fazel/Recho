import { formatFileSize } from '../lib/format'
import type { FileCategory, MessageFile } from '../lib/types'

export type PendingFile = {
  file: File
  category: FileCategory
  /** Local object URL used for the preview; revoked when the draft is cleared. */
  previewUrl: string
  status: 'ready' | 'uploading' | 'uploaded' | 'error'
  progress: number
  error: string
  /** Set once the upload succeeds so a retry reuses it instead of re-uploading. */
  uploaded: MessageFile | null
}

type FileUploadPreviewProps = {
  attachment: PendingFile
  onRemove: () => void
  onRetry: () => void
}

/** Draft attachment shown above the composer while composing a file message. */
export function FileUploadPreview({
  attachment,
  onRemove,
  onRetry,
}: FileUploadPreviewProps) {
  const { file, category, previewUrl, status, progress, error } = attachment
  const name = file.name.trim() || 'File'

  return (
    <div className="attachment">
      <div className="attachment-body">
        {category === 'image' ? (
          <img className="attachment-thumb" src={previewUrl} alt="" />
        ) : category === 'video' ? (
          <video
            className="attachment-thumb"
            src={previewUrl}
            muted
            playsInline
            preload="metadata"
          />
        ) : category === 'voice' ? (
          <div className="attachment-audio">
            <audio src={previewUrl} controls preload="metadata" />
          </div>
        ) : (
          <span className="attachment-icon" aria-hidden>
            <svg viewBox="0 0 24 24">
              <path d="M6 3.5h7L18 8.5v12a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1v-16a1 1 0 0 1 1-1Z" />
              <path d="M13 3.5V8.5H18" />
            </svg>
          </span>
        )}

        <div className="attachment-meta">
          <span className="attachment-name" title={name}>
            {name}
          </span>
          <span className="attachment-size">
            {formatFileSize(file.size)}
            {status === 'uploading' ? ` · Uploading ${progress}%` : null}
            {status === 'uploaded' ? ' · Ready to send' : null}
          </span>

          {status === 'uploading' ? (
            <span
              className="attachment-progress"
              role="progressbar"
              aria-valuemin={0}
              aria-valuemax={100}
              aria-valuenow={progress}
            >
              <span style={{ width: `${progress}%` }} />
            </span>
          ) : null}

          {status === 'error' && error ? (
            <span className="attachment-error" role="alert">
              {error}
            </span>
          ) : null}
        </div>

        {status === 'error' ? (
          <button
            type="button"
            className="attachment-retry"
            onClick={onRetry}
            title="Retry upload"
          >
            Retry
          </button>
        ) : null}

        <button
          type="button"
          className="attachment-remove"
          onClick={onRemove}
          aria-label="Remove attachment"
          title="Remove"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M6 6l12 12M18 6 6 18" />
          </svg>
        </button>
      </div>
    </div>
  )
}
