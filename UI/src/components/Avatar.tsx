import type { CSSProperties } from 'react'
import { displayName, hueFromId, initials } from '../lib/format'

type AvatarProps = {
  id: string
  name: string
  url?: string
  size?: 'sm' | 'md' | 'lg'
}

export function Avatar({ id, name, url, size = 'md' }: AvatarProps) {
  const hue = hueFromId(id || name)

  if (url) {
    return (
      <img
        className={`avatar avatar-${size}`}
        src={url}
        alt=""
        style={{ '--hue': hue } as CSSProperties}
      />
    )
  }

  return (
    <span
      className={`avatar avatar-${size}`}
      style={{ '--hue': hue } as CSSProperties}
      aria-hidden
    >
      {initials(displayName({ display_name: name }))}
    </span>
  )
}
