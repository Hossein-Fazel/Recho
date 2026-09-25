import { AudioMessage } from './AudioMessage'
import { DocumentMessage } from './DocumentMessage'
import { ImageMessage } from './ImageMessage'
import { VideoMessage } from './VideoMessage'
import type { CSSProperties } from 'react'
import type { MessageFile } from '../lib/types'

type FileMessageProps = {
  file: MessageFile
  onOpenImage: (file: MessageFile) => void
  uploadStatus?: 'uploading' | 'sending'
  uploadProgress?: number
  localPreviewUrl?: string
  onCancelUpload?: () => void
}

/** Renders a file message payload according to its backend category. */
export function FileMessage({
  file,
  onOpenImage,
  uploadStatus,
  uploadProgress = 0,
  localPreviewUrl,
  onCancelUpload,
}: FileMessageProps) {
  const content = (() => {
    switch (file.category) {
    case 'image':
      return <ImageMessage file={file} onOpen={onOpenImage} />
    case 'video':
      return <VideoMessage file={file} />
    case 'voice':
      return <AudioMessage file={file} uploading={Boolean(uploadStatus)} />
    default:
      return <DocumentMessage file={file} />
    }
  })()

  if (!uploadStatus) return content

  return (
    <div className="file-uploading">
      {file.category === 'image' && localPreviewUrl ? (
        <img className="file-uploading-preview" src={localPreviewUrl} alt={file.file_name ?? 'Image'} />
      ) : file.category === 'video' && localPreviewUrl ? (
        <video className="file-uploading-preview" src={localPreviewUrl} muted playsInline preload="metadata" />
      ) : (
        content
      )}
      <button
        type="button"
        className="file-upload-progress"
        style={{ '--upload-progress': `${Math.max(0, Math.min(100, uploadProgress))}%` } as CSSProperties}
        onClick={onCancelUpload}
        aria-label={uploadStatus === 'uploading' ? 'Cancel upload' : 'Cancel sending'}
        title={uploadStatus === 'uploading' ? 'Cancel upload' : 'Cancel sending'}
      >
        <span className="file-upload-progress-ring" />
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M18 6 6 18" /></svg>
      </button>
    </div>
  )
}
