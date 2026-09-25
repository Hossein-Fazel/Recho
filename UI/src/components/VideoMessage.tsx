import { useState } from 'react'
import { DocumentMessage } from './DocumentMessage'
import type { MessageFile } from '../lib/types'

type VideoMessageProps = {
  file: MessageFile
}

/** Inline video player using the browser's native controls. */
export function VideoMessage({ file }: VideoMessageProps) {
  const [failed, setFailed] = useState(false)

  const url = file.url?.trim()
  if (!url || failed) {
    return <DocumentMessage file={file} />
  }

  const name = file.file_name?.trim() || 'Video'

  return (
    <div className="file-video" title={name}>
      <video
        src={url}
        controls
        preload="metadata"
        playsInline
        onError={() => setFailed(true)}
      />
    </div>
  )
}
