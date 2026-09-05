import { useEffect, useState } from 'react'
import { copyText } from '../lib/clipboard'
import { inviteLink } from '../lib/invite'

type InviteCodeProps = {
  code: string
}

/**
 * Shows an invite code with buttons to copy the code and its link. Copying
 * falls back to selecting the text when the clipboard API is unavailable
 * (non-HTTPS origins, older browsers).
 */
export function InviteCode({ code }: InviteCodeProps) {
  const [copied, setCopied] = useState('')

  useEffect(() => {
    if (!copied) return
    const timer = window.setTimeout(() => setCopied(''), 1800)
    return () => window.clearTimeout(timer)
  }, [copied])

  async function copy(value: string, label: string) {
    setCopied((await copyText(value)) ? label : '')
  }

  const link = inviteLink(code)

  return (
    <div className="invite-box">
      <p className="invite-label">Invite code</p>

      <div className="invite-code-row">
        <code className="invite-code">{code}</code>
        <button
          type="button"
          className="ghost-btn invite-copy"
          onClick={() => void copy(code, 'code')}
        >
          {copied === 'code' ? 'Copied' : 'Copy'}
        </button>
      </div>

      <div className="invite-link-row">
        <input
          className="invite-link"
          value={link}
          readOnly
          aria-label="Invite link"
          onFocus={(e) => e.currentTarget.select()}
        />
        <button
          type="button"
          className="ghost-btn invite-copy"
          onClick={() => void copy(link, 'link')}
        >
          {copied === 'link' ? 'Copied' : 'Copy link'}
        </button>
      </div>

      <p className="invite-hint">
        Anyone with this code can join the group.
      </p>
    </div>
  )
}
