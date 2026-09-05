import { useMemo } from 'react'
import { splitLinks } from '../lib/linkify'

type LinkedTextProps = {
  text: string
  /**
   * Called instead of navigating when a link is an invite for this deployment,
   * so joining a group doesn't reload the whole app.
   */
  onInvite?: (code: string) => void
}

/** Renders text with any URLs in it turned into real links. */
export function LinkedText({ text, onInvite }: LinkedTextProps) {
  const parts = useMemo(() => splitLinks(text), [text])

  return (
    <>
      {parts.map((part, index) => {
        if (part.kind === 'text') return part.value

        const handledInApp = Boolean(part.inviteCode && onInvite)

        return (
          <a
            key={`${index}:${part.href}`}
            className="linkified"
            href={part.href}
            // Invite links stay in this tab: the click is intercepted, and if
            // it ever isn't, the server serves the app for /join/<code> anyway.
            target={handledInApp ? undefined : '_blank'}
            rel="noreferrer noopener"
            onClick={
              handledInApp
                ? (event) => {
                    event.preventDefault()
                    onInvite?.(part.inviteCode)
                  }
                : undefined
            }
          >
            {part.value}
          </a>
        )
      })}
    </>
  )
}
