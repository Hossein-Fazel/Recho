import { useEffect, useState } from 'react'
import { Avatar } from './Avatar'
import { LinkedText } from './LinkedText'
import { api } from '../lib/api'
import { handle, personName } from '../lib/format'
import type { User } from '../lib/types'

type UserInfoPanelProps = {
  userId: string | null
  open: boolean
  onClose: () => void
}

const EMPTY_PLACEHOLDER = 'Not set'

type InfoRowProps = {
  label: string
  value: string
}

function InfoRow({ label, value }: InfoRowProps) {
  const filled = value.trim().length > 0

  return (
    <div className="info-row">
      <dt className="info-label">{label}</dt>
      <dd className={`info-value${filled ? '' : ' info-value-empty'}`}>
        {filled ? <LinkedText text={value} /> : EMPTY_PLACEHOLDER}
      </dd>
    </div>
  )
}

export function UserInfoPanel({ userId, open, onClose }: UserInfoPanelProps) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open || !userId) return

    const targetUserId = userId
    let cancelled = false

    async function loadUser() {
      setLoading(true)
      setError('')
      setUser(null)

      try {
        const result = await api.getUser(targetUserId)
        if (!cancelled) setUser(result)
      } catch {
        if (!cancelled) setError('Could not load user profile')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void loadUser()

    return () => {
      cancelled = true
    }
  }, [open, userId])

  useEffect(() => {
    if (!open) return

    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!open || !userId) return null

  const title = user ? personName(user) : 'Profile'
  const username = user?.username?.trim() ?? ''
  const avatarId = user?.id || userId

  return (
    <>
      <button
        type="button"
        className="info-backdrop"
        onClick={onClose}
        aria-label="Close profile"
        tabIndex={-1}
      />

      <aside
        className="info-panel"
        role="dialog"
        aria-modal="true"
        aria-label="User profile"
      >
        <header className="info-head">
          <button
            type="button"
            className="icon-button info-close"
            onClick={onClose}
            aria-label="Close profile"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M6 6l12 12M18 6 6 18" />
            </svg>
          </button>
          <span className="info-head-label">Profile</span>
        </header>

        <div className="info-scroll">
          <section className="info-hero">
            <Avatar
              id={avatarId}
              name={title}
              url={user?.avatar_url || undefined}
              size="xl"
            />
            <h2>{title}</h2>
            <p className="info-subtitle">
              {handle(username) || 'User'}
            </p>
          </section>

          {loading ? <p className="info-note">Loading…</p> : null}
          {error ? <p className="info-note error">{error}</p> : null}

          {user ? (
            <dl className="info-card">
              <InfoRow label="Name" value={user.display_name} />
              <InfoRow label="Username" value={handle(user.username)} />
              <InfoRow label="Bio" value={user.bio} />
            </dl>
          ) : null}
        </div>
      </aside>
    </>
  )
}
