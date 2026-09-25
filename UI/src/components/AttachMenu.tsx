import { memo, useEffect, useRef, type RefObject } from 'react'
import { FileTypeIcon } from './FileTypeIcon'

/** Logical attachment categories offered by the `+` menu. */
export type AttachOption = 'media' | 'audio' | 'file'

type AttachMenuProps = {
  open: boolean
  triggerRef: RefObject<HTMLButtonElement | null>
  onSelect: (option: AttachOption) => void
  onClose: () => void
}

/**
 * Small popover opened from the composer's `+` button, offering the three
 * attachment categories (photo/video, audio, generic file). Mirrors
 * EmojiPicker's outside-press/Escape handling so it behaves like the rest of
 * the composer's popovers.
 */
function AttachMenuInner({
  open,
  triggerRef,
  onSelect,
  onClose,
}: AttachMenuProps) {
  const panelRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return

    function onPointerDown(event: PointerEvent) {
      const target = event.target as Node
      if (panelRef.current?.contains(target)) return
      if (triggerRef.current?.contains(target)) return
      onClose()
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }

    // `pointerdown` rather than `mousedown`, matching EmojiPicker: a touch
    // that lands on a scrollable area never dispatches synthetic mouse
    // events, which would leave the menu stuck open on a phone.
    document.addEventListener('pointerdown', onPointerDown)
    document.addEventListener('keydown', onKeyDown)

    return () => {
      document.removeEventListener('pointerdown', onPointerDown)
      document.removeEventListener('keydown', onKeyDown)
    }
  }, [open, onClose, triggerRef])

  if (!open) return null

  return (
    <div
      ref={panelRef}
      className="attach-menu-wrap"
      role="menu"
      aria-label="Attach"
    >
      <button
        type="button"
        className="attach-menu-item"
        role="menuitem"
        onClick={() => onSelect('media')}
      >
        <span className="attach-menu-icon" aria-hidden="true">
          <FileTypeIcon category="image" />
        </span>
        Photo &amp; Video
      </button>

      <button
        type="button"
        className="attach-menu-item"
        role="menuitem"
        onClick={() => onSelect('audio')}
      >
        <span className="attach-menu-icon" aria-hidden="true">
          <FileTypeIcon category="voice" />
        </span>
        Audio
      </button>

      <button
        type="button"
        className="attach-menu-item"
        role="menuitem"
        onClick={() => onSelect('file')}
      >
        <span className="attach-menu-icon" aria-hidden="true">
          <FileTypeIcon />
        </span>
        File
      </button>
    </div>
  )
}

export const AttachMenu = memo(AttachMenuInner)
