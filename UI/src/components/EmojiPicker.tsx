import {
  memo,
  useEffect,
  useRef,
  useState,
  type RefObject,
} from 'react'
import EmojiPickerReact, {
  EmojiStyle,
  Theme,
} from 'emoji-picker-react'

type EmojiPickerProps = {
  open: boolean
  triggerRef: RefObject<HTMLButtonElement | null>
  onSelect: (emoji: string) => void
  onClose: () => void
}

function EmojiPickerInner({
  open,
  triggerRef,
  onSelect,
  onClose,
}: EmojiPickerProps) {
  const panelRef = useRef<HTMLDivElement>(null)
  const [theme, setTheme] = useState<'dark' | 'light'>(() =>
    document.documentElement.dataset.theme === 'light'
      ? 'light'
      : 'dark',
  )

  useEffect(() => {
    if (!open) return

    const root = document.documentElement
    const updateTheme = () =>
      setTheme(
        root.dataset.theme === 'light' ? 'light' : 'dark',
      )
    const observer = new MutationObserver(updateTheme)
    observer.observe(root, {
      attributes: true,
      attributeFilter: ['data-theme'],
    })

    function onPointerDown(event: MouseEvent) {
      const target = event.target as Node
      if (panelRef.current?.contains(target)) return
      if (triggerRef.current?.contains(target)) return
      onClose()
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    document.addEventListener('mousedown', onPointerDown)
    document.addEventListener('keydown', onKeyDown)

    return () => {
      observer.disconnect()
      document.removeEventListener('mousedown', onPointerDown)
      document.removeEventListener('keydown', onKeyDown)
    }
  }, [open, onClose, triggerRef])

  if (!open) return null

  return (
    <div
      ref={panelRef}
      className="emoji-picker-wrap"
      role="dialog"
      aria-label="Emoji picker"
    >
      <EmojiPickerReact
        width="100%"
        height={380}
        theme={theme === 'light' ? Theme.LIGHT : Theme.DARK}
        emojiStyle={EmojiStyle.NATIVE}
        lazyLoadEmojis
        autoFocusSearch={false}
        skinTonesDisabled
        searchPlaceHolder="Search emojis"
        previewConfig={{ showPreview: false }}
        className="epr-theme-app"
        onEmojiClick={(clicked) => onSelect(clicked.emoji)}
      />
    </div>
  )
}

export const EmojiPicker = memo(EmojiPickerInner)