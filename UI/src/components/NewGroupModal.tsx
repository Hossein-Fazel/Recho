import { useEffect, useRef, useState } from 'react'
import { api } from '../lib/api'
import { InviteCode } from './InviteCode'
import type { CreateGroupResponse } from '../lib/types'

type NewGroupModalProps = {
  open: boolean
  onClose: () => void
  onCreated: (group: CreateGroupResponse) => void
}

const NAME_LIMIT = 100
const BIO_LIMIT = 250

export function NewGroupModal({ open, onClose, onCreated }: NewGroupModalProps) {
  const [name, setName] = useState('')
  const [bio, setBio] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [created, setCreated] = useState<CreateGroupResponse | null>(null)

  const nameInput = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!open) return

    setName('')
    setBio('')
    setError('')
    setSaving(false)
    setCreated(null)
    nameInput.current?.focus()
  }, [open])

  useEffect(() => {
    if (!open) return

    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!open) return null

  async function submit(event: React.FormEvent) {
    event.preventDefault()

    const trimmed = name.trim()
    if (!trimmed) {
      setError('Give the group a name')
      return
    }

    setSaving(true)
    setError('')

    try {
      const group = await api.createGroup(trimmed, bio.trim())
      setCreated(group)
      onCreated(group)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create the group')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-overlay" role="dialog" aria-modal="true">
      <div className="modal">
        {created ? (
          <>
            <h3>{created.name} is ready</h3>
            <p>
              Share this invite code and anyone can join the group.
            </p>

            <InviteCode code={created.invite_code} />

            <div className="modal-actions">
              <button type="button" className="primary compact" onClick={onClose}>
                Done
              </button>
            </div>
          </>
        ) : (
          <form className="auth-form" onSubmit={(e) => void submit(e)}>
            <h3>New group</h3>

            <label>
              Name
              <input
                ref={nameInput}
                value={name}
                maxLength={NAME_LIMIT}
                onChange={(e) => setName(e.target.value)}
                placeholder="Weekend plans"
                aria-label="Group name"
              />
            </label>

            <label>
              About <span className="label-hint">optional</span>
              <textarea
                value={bio}
                rows={3}
                maxLength={BIO_LIMIT}
                onChange={(e) => setBio(e.target.value)}
                placeholder="What is this group for?"
                aria-label="Group description"
              />
            </label>

            {error ? <p className="form-error">{error}</p> : null}

            <div className="modal-actions">
              <button
                type="button"
                className="ghost-btn"
                onClick={onClose}
                disabled={saving}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="primary compact"
                disabled={saving || !name.trim()}
              >
                {saving ? 'Creating…' : 'Create group'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}
