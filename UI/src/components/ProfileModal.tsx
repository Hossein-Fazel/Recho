import { useEffect, useRef, useState } from 'react'
import { useAuth } from '../context/AuthContext'
import { personName } from '../lib/format'
import { Avatar } from './Avatar'

type ProfileModalProps = {
  onClose: () => void
}

const NAME_LIMIT = 100
const BIO_LIMIT = 250
const MAX_AVATAR_BYTES = 5 * 1024 * 1024
const ACCEPTED_IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/webp', 'image/gif']

/**
 * Telegram-style profile editor: the avatar is changed by tapping the photo,
 * while the name and bio are edited inline and saved together. Mounted only
 * while open so the fields always start from the latest profile.
 */
export function ProfileModal({ onClose }: ProfileModalProps) {
  const { user, updateProfile, updateAvatar } = useAuth()

  const [displayName, setDisplayName] = useState(() => user?.display_name ?? '')
  const [bio, setBio] = useState(() => user?.bio ?? '')
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const [preview, setPreview] = useState('')

  const fileInput = useRef<HTMLInputElement>(null)

  useEffect(() => {
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])

  // Object URLs leak if they aren't revoked once the preview is replaced.
  useEffect(() => {
    if (!preview) return
    return () => URL.revokeObjectURL(preview)
  }, [preview])

  if (!user) return null

  const name = personName(user)
  const trimmedName = displayName.trim()
  const trimmedBio = bio.trim()
  const currentName = (user.display_name ?? '').trim()
  const currentBio = (user.bio ?? '').trim()

  const nameDirty = trimmedName !== currentName
  const bioDirty = trimmedBio !== currentBio
  const dirty = nameDirty || bioDirty
  const busy = saving || uploading

  async function pickAvatar(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    // Reset so picking the same file again still fires onChange.
    event.target.value = ''
    if (!file) return

    if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
      setError('Choose a PNG, JPEG, WEBP or GIF image')
      return
    }

    if (file.size > MAX_AVATAR_BYTES) {
      setError('Image must be smaller than 5 MB')
      return
    }

    setError('')
    setPreview(URL.createObjectURL(file))
    setUploading(true)

    try {
      await updateAvatar(file)
    } catch (err) {
      setPreview('')
      setError(err instanceof Error ? err.message : 'Could not update your photo')
    } finally {
      setUploading(false)
    }
  }

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    if (busy || !dirty) return

    if (nameDirty && !trimmedName) {
      setError('Name can’t be empty')
      return
    }

    setSaving(true)
    setError('')

    try {
      const patch: { display_name?: string; bio?: string } = {}
      if (nameDirty) patch.display_name = trimmedName
      if (bioDirty) patch.bio = trimmedBio

      await updateProfile(patch)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save your profile')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-overlay" role="dialog" aria-modal="true">
      <div className="modal profile-modal">
        <header className="profile-head">
          <h3>Edit profile</h3>
          <button
            type="button"
            className="icon-button profile-close"
            onClick={onClose}
            disabled={busy}
            aria-label="Close"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M6 6l12 12M18 6 6 18" />
            </svg>
          </button>
        </header>

        <form className="auth-form profile-form" onSubmit={(e) => void submit(e)}>
          <div className="profile-photo">
            <button
              type="button"
              className="profile-avatar"
              onClick={() => fileInput.current?.click()}
              disabled={busy}
              aria-label="Change profile photo"
              title="Change profile photo"
            >
              <Avatar
                id={user.id}
                name={name}
                url={preview || user.avatar_url || undefined}
                size="xl"
              />
              <span className="profile-avatar-overlay" aria-hidden>
                {uploading ? (
                  <span className="spinner" />
                ) : (
                  <svg viewBox="0 0 24 24">
                    <path d="M4 8.5A1.5 1.5 0 0 1 5.5 7h1.7l1.1-1.6A1.5 1.5 0 0 1 9.55 5h4.9a1.5 1.5 0 0 1 1.25.4L16.8 7h1.7A1.5 1.5 0 0 1 20 8.5v9A1.5 1.5 0 0 1 18.5 19h-13A1.5 1.5 0 0 1 4 17.5z" />
                    <circle cx="12" cy="12.5" r="3.3" />
                  </svg>
                )}
              </span>
            </button>

            <span className="profile-photo-hint">
              {uploading ? 'Uploading…' : 'Tap the photo to change it'}
            </span>

            <input
              ref={fileInput}
              type="file"
              accept={ACCEPTED_IMAGE_TYPES.join(',')}
              onChange={(e) => void pickAvatar(e)}
              hidden
            />
          </div>

          <label>
            Name
            <input
              value={displayName}
              maxLength={NAME_LIMIT}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Your name"
              autoComplete="name"
            />
          </label>

          <label>
            Bio <span className="label-hint">optional</span>
            <textarea
              value={bio}
              rows={3}
              maxLength={BIO_LIMIT}
              onChange={(e) => setBio(e.target.value)}
              placeholder="A few words about you"
            />
          </label>
          <span className="profile-counter">
            {bio.length}/{BIO_LIMIT}
          </span>

          <label>
            Username
            <input value={`@${user.username}`} readOnly aria-label="Username" />
          </label>

          {error ? <p className="form-error">{error}</p> : null}

          <div className="modal-actions">
            <button
              type="button"
              className="ghost-btn"
              onClick={onClose}
              disabled={busy}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="primary compact"
              disabled={busy || !dirty}
            >
              {saving ? 'Saving…' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
