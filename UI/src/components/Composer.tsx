import { useEffect, useRef, useState } from 'react'

type ComposerProps = {
  disabled: boolean
  onSend: (content: string) => void
}

export function Composer({ disabled, onSend }: ComposerProps) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    const textarea = textareaRef.current
    if (!textarea) return
    textarea.style.height = '0px'
    textarea.style.height = `${Math.min(textarea.scrollHeight, 144)}px`
  }, [value])

  function submit() {
    const content = value.trim()
    if (!content || disabled) return
    onSend(content)
    setValue('')
  }

  return (
    <form
      className="composer"
      onSubmit={(event) => {
        event.preventDefault()
        submit()
      }}
    >
      <div className="composer-box">
        <textarea
          ref={textareaRef}
          rows={1}
          placeholder={disabled ? 'Choose a conversation' : 'Write a message…'}
          value={value}
          disabled={disabled}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault()
              submit()
            }
          }}
          aria-label="Message"
        />
        <button
          className="send"
          type="submit"
          disabled={disabled || !value.trim()}
          aria-label="Send message"
          title="Send message"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="m5 12 14-7-4 14-3.2-5.8L5 12Z" />
            <path d="M11.8 13.2 19 5" />
          </svg>
        </button>
      </div>
      {!disabled ? <span className="composer-hint">Enter to send · Shift + Enter for a new line</span> : null}
    </form>
  )
}
