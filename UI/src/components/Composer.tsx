import {
  useCallback,
  useEffect,
  useRef,
  useState,
} from 'react'
import { EmojiPicker } from './EmojiPicker'

type ComposerProps = {
  disabled: boolean
  onSend: (content: string) => void
}

export function Composer({ disabled, onSend }: ComposerProps) {
  const [value, setValue] = useState('')
  const [pickerOpen, setPickerOpen] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const emojiToggleRef = useRef<HTMLButtonElement>(null)
  const valueRef = useRef('')
  const cursorPos = useRef(0)

  function setMessage(next: string) {
    valueRef.current = next
    setValue(next)
  }

  useEffect(() => {
    const textarea = textareaRef.current
    if (!textarea) return
    textarea.style.height = '0px'
    textarea.style.height = `${Math.min(textarea.scrollHeight, 144)}px`
  }, [value])

  const rememberCursor = useCallback(() => {
    const textarea = textareaRef.current
    if (textarea) {
      cursorPos.current =
        textarea.selectionStart ?? cursorPos.current
    }
  }, [])

  const onSelect = useCallback((emoji: string) => {
    rememberCursor()
    const pos = cursorPos.current
    const current = valueRef.current
    const next =
      current.slice(0, pos) + emoji + current.slice(pos)
    setMessage(next)

    const after = pos + emoji.length
    cursorPos.current = after
    requestAnimationFrame(() => {
      const textarea = textareaRef.current
      textarea?.focus()
      textarea?.setSelectionRange(after, after)
    })
  }, [rememberCursor])

  const closePicker = useCallback(() => setPickerOpen(false), [])

  function submit() {
    const content = valueRef.current.trim()
    if (!content || disabled) return
    onSend(content)
    setMessage('')
    cursorPos.current = 0
    setPickerOpen(false)
  }

  return (
    <div className="composer">
      <form
        className="composer-form"
        onSubmit={(event) => {
          event.preventDefault()
          submit()
        }}
      >
        <div className="composer-box">
          {!disabled ? (
            <button
              ref={emojiToggleRef}
              type="button"
              className="emoji-toggle"
              onClick={() => setPickerOpen((open) => !open)}
              aria-label="Emoji picker"
              aria-expanded={pickerOpen}
              title="Emoji picker"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="12" cy="12" r="9" />
                <path d="M8.5 14.2s1.1 1.5 3.5 1.5 3.5-1.5 3.5-1.5" />
                <path d="M9 9.6h.01M15 9.6h.01" />
              </svg>
            </button>
          ) : null}
          <textarea
            ref={textareaRef}
            rows={1}
            placeholder={disabled ? 'Choose a conversation' : 'Write a message…'}
            value={value}
            disabled={disabled}
            onChange={(e) => {
              setMessage(e.target.value)
              cursorPos.current =
                e.target.selectionStart ?? cursorPos.current
            }}
            onKeyUp={rememberCursor}
            onClick={rememberCursor}
            onSelect={rememberCursor}
            onFocus={rememberCursor}
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
      <EmojiPicker
        open={pickerOpen && !disabled}
        triggerRef={emojiToggleRef}
        onSelect={onSelect}
        onClose={closePicker}
      />
    </div>
  )
}