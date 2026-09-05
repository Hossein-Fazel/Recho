import { useCallback, useEffect, useRef, useState } from 'react'
import { Avatar } from './Avatar'
import { LinkedText } from './LinkedText'
import { api } from '../lib/api'
import { UNNAMED_GROUP } from '../lib/format'
import type { GroupPreview } from '../lib/types'

type JoinGroupModalProps = {
  open: boolean
  /** Code taken from an invite link. Empty when the user opened this manually. */
  initialCode?: string
  onClose: () => void
  /** Called once the user is a member, with the group's conversation id. */
  onJoined: (conversationId: string) => void
}

/**
 * Shows the group behind an invite code, then a Join button. Handles both
 * arriving from an invite link (code prefilled, looked up immediately) and
 * pasting a code by hand.
 */
export function JoinGroupModal({
  open,
  initialCode = '',
  onClose,
  onJoined,
}: JoinGroupModalProps) {
  const [code, setCode] = useState(initialCode)
  const [preview, setPreview] = useState<GroupPreview | null>(null)
  const [looking, setLooking] = useState(false)
  const [joining, setJoining] = useState(false)
  const [error, setError] = useState('')

  const codeInput = useRef<HTMLInputElement>(null)

  const lookup = useCallback(async (value: string) => {
    const trimmed = value.trim()
    if (!trimmed) {
      setError('Enter an invite code')
      return
    }

    setLooking(true)
    setError('')
    setPreview(null)

    try {
      setPreview(await api.groupByInviteCode(trimmed))
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'That invite code is not valid',
      )
    } finally {
      setLooking(false)
    }
  }, [])

  useEffect(() => {
    if (!open) return

    setCode(initialCode)
    setPreview(null)
    setError('')
    setJoining(false)

    if (initialCode.trim()) {
      void lookup(initialCode)
    } else {
      codeInput.current?.focus()
    }
  }, [open, initialCode, lookup])

  useEffect(() => {
    if (!open) return

    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!open) return null

  async function join() {
    if (!preview) return

    // Already a member: just open the chat.
    if (preview.is_member) {
      onJoined(preview.conversation_id)
      return
    }

    setJoining(true)
    setError('')

    try {
      const joined = await api.joinGroup(code.trim())
      onJoined(joined.conversation_id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not join the group')
    } finally {
      setJoining(false)
    }
  }

  const groupName = preview?.name.trim() || UNNAMED_GROUP
  const memberLabel = preview
    ? `${preview.member_count} member${preview.member_count === 1 ? '' : 's'}`
    : ''

  return (
    <div className="modal-overlay" role="dialog" aria-modal="true">
      <div className="modal">
        <h3>Join a group</h3>

        {preview ? (
          <>
            <div className="join-preview">
              {!preview.is_member ? (
                <p className="join-intro">
                  You’ve been invited to join this group
                </p>
              ) : null}
              <Avatar
                id={preview.conversation_id}
                name={groupName}
                url={preview.avatar_url || undefined}
                size="xl"
              />
              <h4 className="join-name">{groupName}</h4>
              <p className="join-meta">{memberLabel}</p>
              {preview.bio.trim() ? (
                <p className="join-bio">
                  <LinkedText text={preview.bio} />
                </p>
              ) : null}
            </div>

            {error ? <p className="form-error">{error}</p> : null}

            <div className="modal-actions">
              <button
                type="button"
                className="ghost-btn"
                onClick={onClose}
                disabled={joining}
              >
                Cancel
              </button>
              <button
                type="button"
                className="primary compact"
                onClick={() => void join()}
                disabled={joining}
              >
                {preview.is_member
                  ? 'Open chat'
                  : joining
                    ? 'Joining…'
                    : 'Join'}
              </button>
            </div>
          </>
        ) : (
          <form
            className="auth-form"
            onSubmit={(e) => {
              e.preventDefault()
              void lookup(code)
            }}
          >
            <p className="join-intro">
              Paste the invite code someone shared with you.
            </p>

            <label>
              Invite code
              <input
                ref={codeInput}
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="ABCD234XYZ"
                aria-label="Invite code"
                autoComplete="off"
                spellCheck={false}
              />
            </label>

            {error ? <p className="form-error">{error}</p> : null}

            <div className="modal-actions">
              <button
                type="button"
                className="ghost-btn"
                onClick={onClose}
                disabled={looking}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="primary compact"
                disabled={looking || !code.trim()}
              >
                {looking ? 'Looking…' : 'Find group'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}
