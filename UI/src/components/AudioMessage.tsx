import { useEffect, useRef, useState } from 'react'
import type { CSSProperties } from 'react'
import { DocumentMessage } from './DocumentMessage'
import type { MessageFile } from '../lib/types'

type AudioMessageProps = {
  file: MessageFile
  uploading?: boolean
}

/** `m:ss` (or `h:mm:ss`) clock. Returns `--:--` for unknown durations. */
function formatClock(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return '--:--'

  const total = Math.floor(seconds)
  const secs = total % 60
  const mins = Math.floor(total / 60) % 60
  const hours = Math.floor(total / 3600)
  const mm = hours > 0 ? String(mins).padStart(2, '0') : String(mins)

  return hours > 0
    ? `${hours}:${mm}:${String(secs).padStart(2, '0')}`
    : `${mm}:${String(secs).padStart(2, '0')}`
}

const PlayIcon = (
  <svg viewBox="0 0 24 24" aria-hidden="true">
    <path d="M8 5.4v13.2L18.6 12 8 5.4Z" />
  </svg>
)

const PauseIcon = (
  <svg viewBox="0 0 24 24" aria-hidden="true">
    <rect x="7" y="5" width="3.6" height="14" rx="1.3" />
    <rect x="13.4" y="5" width="3.6" height="14" rx="1.3" />
  </svg>
)

/**
 * Compact music/audio player for file messages whose backend category is
 * `voice`. Unlike a voice note, it shows the real filename, current time,
 * duration, and a seek bar. Playback state is fully local so several audio
 * messages in one conversation never interfere with each other.
 */
export function AudioMessage({ file, uploading = false }: AudioMessageProps) {
  const audioRef = useRef<HTMLAudioElement | null>(null)
  const [failed, setFailed] = useState(false)
  const [playing, setPlaying] = useState(false)
  const [duration, setDuration] = useState(0)
  const [currentTime, setCurrentTime] = useState(0)

  const url = file.url?.trim()
  const name = file.file_name?.trim() || 'Audio'
  const hasDuration = Number.isFinite(duration) && duration > 0
  const progress = hasDuration
    ? Math.min(100, Math.max(0, (currentTime / duration) * 100))
    : 0

  // Reset the player when the source changes (e.g. an edited message).
  useEffect(() => {
    setFailed(false)
    setPlaying(false)
    setDuration(0)
    setCurrentTime(0)
  }, [url])

  // Stop playback if the source changes or the message unmounts.
  useEffect(() => {
    const audio = audioRef.current
    return () => {
      audio?.pause()
    }
  }, [url])

  if (!url && uploading) {
    return (
      <div className="audio-player audio-player-uploading" title={name}>
        <button type="button" className="audio-player-toggle" disabled aria-label={`Uploading ${name}`}>{PlayIcon}</button>
        <div className="audio-player-body">
          <div className="audio-player-head">
            <span className="audio-player-name" title={name}>{name}</span>
            <span className="audio-player-time">Uploading…</span>
          </div>
          <div className="audio-player-seek audio-player-seek-disabled" />
        </div>
      </div>
    )
  }

  if (!url || failed) {
    return <DocumentMessage file={file} />
  }

  function togglePlayback() {
    const audio = audioRef.current
    if (!audio) return

    if (audio.paused) {
      // `play()` rejects when playback is blocked; fall back to the document
      // card rather than leaving a dead control.
      void audio.play().catch(() => setFailed(true))
    } else {
      audio.pause()
    }
  }

  function seek(next: number) {
    const audio = audioRef.current
    if (!audio || !hasDuration) return

    audio.currentTime = next
    setCurrentTime(next)
  }

  return (
    <div className="audio-player" title={name}>
      <button
        type="button"
        className="audio-player-toggle"
        onClick={togglePlayback}
        aria-label={playing ? `Pause ${name}` : `Play ${name}`}
        title={playing ? 'Pause' : 'Play'}
      >
        {playing ? PauseIcon : PlayIcon}
      </button>

      <div className="audio-player-body">
        <div className="audio-player-head">
          <span className="audio-player-name" title={name}>
            {name}
          </span>
          <span className="audio-player-time">
            {formatClock(currentTime)} / {hasDuration ? formatClock(duration) : '--:--'}
          </span>
        </div>

        <input
          type="range"
          className="audio-player-seek"
          min={0}
          max={hasDuration ? duration : 0}
          step="any"
          value={hasDuration ? Math.min(currentTime, duration) : 0}
          onChange={(event) => seek(Number(event.target.value))}
          disabled={!hasDuration}
          aria-label={`Seek ${name}`}
          style={{ '--audio-progress': `${progress}%` } as CSSProperties}
        />
      </div>

      <audio
        ref={audioRef}
        src={url}
        preload="metadata"
        onLoadedMetadata={(event) => setDuration(event.currentTarget.duration)}
        onDurationChange={(event) => setDuration(event.currentTarget.duration)}
        onTimeUpdate={(event) => setCurrentTime(event.currentTarget.currentTime)}
        onPlay={() => setPlaying(true)}
        onPlaying={() => setPlaying(true)}
        onPause={() => setPlaying(false)}
        onEnded={() => {
          setPlaying(false)
          setCurrentTime(0)
        }}
        onError={() => setFailed(true)}
      />
    </div>
  )
}
