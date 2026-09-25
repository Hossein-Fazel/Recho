import { useCallback, useEffect, useRef, useState } from 'react'
import { AttachMenu, type AttachOption } from './AttachMenu'
import { EmojiPicker } from './EmojiPicker'
import type { FileCategory } from '../lib/types'
import { validateFile } from '../lib/media'

const ATTACH_ACCEPT: Record<AttachOption, string> = {
  media: 'image/*,video/*',
  audio: 'audio/*',
  file: '',
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

type ComposerProps = {
  disabled: boolean
  conversationId: string | null
  onSend: (content: string) => void
  onFileSelected: (file: File, category: FileCategory) => string | null
  onFileSend: (clientMessageId: string) => void
  onFileCaptionChange: (clientMessageId: string, caption: string) => void
  selectedFile: { file: File; category: FileCategory } | null
  fileSelected: boolean
  editing: { id: number; content: string; file: import('../lib/types').MessageFile | null } | null
  onCancelEdit: () => void
}

export function Composer({
  disabled,
  conversationId,
  onSend,
  onFileSelected,
  onFileSend,
  onFileCaptionChange,
  selectedFile,
  fileSelected,
  editing,
  onCancelEdit,
}: ComposerProps) {
  const [value, setValue] = useState('')
  const [pickerOpen, setPickerOpen] = useState(false)
  const [attachMenuOpen, setAttachMenuOpen] = useState(false)
  const [activeFileId, setActiveFileId] = useState<string | null>(null)
  const [error, setError] = useState('')

  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const emojiToggleRef = useRef<HTMLButtonElement>(null)
  const attachToggleRef = useRef<HTMLButtonElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const valueRef = useRef('')
  const cursorPos = useRef(0)

  function setMessage(next: string) {
    valueRef.current = next
    setValue(next)
    if (activeFileId) onFileCaptionChange(activeFileId, next)
  }

  // Clears the textarea without notifying onFileCaptionChange. Using
  // setMessage('') for this would re-fire the caption callback with an empty
  // string (activeFileId state hasn't cleared yet in the same tick),
  // wiping out the caption we just sent in onFileSend.
  function resetComposerText() {
    valueRef.current = ''
    setValue('')
  }

  useEffect(() => {
    valueRef.current = ''
    setValue('')
    setActiveFileId(null)
    setError('')
  }, [conversationId, editing?.id])

  useEffect(() => {
    const textarea = textareaRef.current
    if (!textarea) return
    textarea.style.height = '0px'
    textarea.style.height = `${Math.min(textarea.scrollHeight, 144)}px`
  }, [value])

  const [selectedFilePreviewUrl, setSelectedFilePreviewUrl] = useState<string | null>(null)

  useEffect(() => {
    if (!selectedFile) {
      setSelectedFilePreviewUrl(null)
      return
    }

    const url = URL.createObjectURL(selectedFile.file)
    setSelectedFilePreviewUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [selectedFile])

  const rememberCursor = useCallback(() => {
    const textarea = textareaRef.current
    if (textarea) cursorPos.current = textarea.selectionStart ?? cursorPos.current
  }, [])

  const onSelect = useCallback((emoji: string) => {
    rememberCursor()
    const pos = cursorPos.current
    const next = valueRef.current.slice(0, pos) + emoji + valueRef.current.slice(pos)
    setMessage(next)
    const after = pos + emoji.length
    cursorPos.current = after
    requestAnimationFrame(() => {
      textareaRef.current?.focus()
      textareaRef.current?.setSelectionRange(after, after)
    })
  }, [rememberCursor, activeFileId])

  const closePicker = useCallback(() => setPickerOpen(false), [])
  const closeAttachMenu = useCallback(() => setAttachMenuOpen(false), [])

  function selectAttachOption(option: AttachOption) {
    setAttachMenuOpen(false)
    const input = fileInputRef.current
    if (input) input.accept = ATTACH_ACCEPT[option]
    input?.click()
  }

  function chooseFile(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) return

    const result = validateFile(file)
    if (!result.ok) {
      setError(result.error)
      return
    }

    setError('')
    const clientMessageId = onFileSelected(file, result.category)
    if (!clientMessageId) return

    setActiveFileId(clientMessageId)
    setMessage('')
    cursorPos.current = 0
    requestAnimationFrame(() => textareaRef.current?.focus())
  }

  function submit() {
    if (disabled) return

    if (editing) {
      const content = valueRef.current.trim()
      if (!content) return
      onSend(content)
      setMessage('')
      setActiveFileId(null)
      cursorPos.current = 0
      return
    }

    if (fileSelected && activeFileId) {
      onFileSend(activeFileId)
      resetComposerText()
      setActiveFileId(null)
      cursorPos.current = 0
      setPickerOpen(false)
      return
    }

    const content = valueRef.current.trim()
    if (!content) return
    onSend(content)
    setMessage('')
    cursorPos.current = 0
    setPickerOpen(false)
  }

  function cancelEdit() {
    setMessage('')
    setActiveFileId(null)
    cursorPos.current = 0
    onCancelEdit()
  }

  const placeholder = disabled
    ? 'Choose a conversation'
    : editing
      ? editing.file ? 'Edit caption…' : 'Edit message…'
      : fileSelected ? 'Add a caption…' : 'Write a message…'

  return (
    <div className="composer">
      {editing ? (
        <div className="edit-banner">
          <button type="button" className="edit-banner-cancel" onClick={cancelEdit} aria-label="Cancel edit" title="Cancel edit">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M18 6 6 18" /></svg>
          </button>
          <div className="edit-banner-text">
            <span className="edit-banner-label">{editing.file ? 'Editing caption' : 'Editing message'}</span>
            <span className="edit-banner-preview">{editing.content || (editing.file ? 'No caption' : '')}</span>
          </div>
        </div>
      ) : null}

      {error ? <p className="composer-error" role="alert">{error}</p> : null}

      {fileSelected && selectedFile ? (
        <div className="composer-file-preview">
          {selectedFile.category === 'image' && selectedFilePreviewUrl ? (
            <img src={selectedFilePreviewUrl} alt={selectedFile.file.name} className="composer-file-preview-media" />
          ) : selectedFile.category === 'video' && selectedFilePreviewUrl ? (
            <video src={selectedFilePreviewUrl} className="composer-file-preview-media" muted playsInline preload="metadata" />
          ) : selectedFile.category === 'voice' && selectedFilePreviewUrl ? (
            <audio src={selectedFilePreviewUrl} controls className="composer-file-preview-audio" />
          ) : (
            <div className="composer-file-preview-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24"><path d="M6 3h8l4 4v14H6z" /><path d="M14 3v5h5M9 13h6M9 17h6" /></svg>
            </div>
          )}
          <div className="composer-file-preview-info">
            <span className="composer-file-preview-name" title={selectedFile.file.name}>{selectedFile.file.name}</span>
            <span className="composer-file-preview-size">{formatFileSize(selectedFile.file.size)}</span>
          </div>
        </div>
      ) : null}

      <form className="composer-form" onSubmit={(event) => { event.preventDefault(); submit() }}>
        <div className="composer-box">
          {!disabled && !editing ? (
            <button ref={attachToggleRef} type="button" className="attach-toggle" onClick={() => { setAttachMenuOpen((open) => !open); setPickerOpen(false) }} disabled={disabled} aria-label="Attach" aria-haspopup="menu" aria-expanded={attachMenuOpen} title="Attach">
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14" /></svg>
            </button>
          ) : null}

          {!disabled ? (
            <button ref={emojiToggleRef} type="button" className="emoji-toggle" onClick={() => { setPickerOpen((open) => !open); setAttachMenuOpen(false) }} aria-label="Emoji picker" aria-expanded={pickerOpen} title="Emoji picker">
              <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M8.5 14.2s1.1 1.5 3.5 1.5 3.5-1.5 3.5-1.5" /><path d="M9 9.6h.01M15 9.6h.01" /></svg>
            </button>
          ) : null}

          <textarea
            ref={textareaRef}
            rows={1}
            placeholder={placeholder}
            value={value}
            disabled={disabled}
            onChange={(e) => { setMessage(e.target.value); cursorPos.current = e.target.selectionStart ?? cursorPos.current }}
            onKeyUp={rememberCursor}
            onClick={rememberCursor}
            onSelect={rememberCursor}
            onFocus={rememberCursor}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); submit() }
              if (e.key === 'Escape' && editing) { e.preventDefault(); cancelEdit() }
            }}
            aria-label="Message"
          />

          <button className="send" type="submit" disabled={disabled || (!value.trim() && !fileSelected && !editing)} onMouseDown={(e) => e.preventDefault()} aria-label={fileSelected ? 'Caption file' : editing ? 'Save edit' : 'Send message'} title={fileSelected ? 'Caption file' : editing ? 'Save edit' : 'Send message'}>
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m5 12 14-7-4 14-3.2-5.8L5 12Z" /><path d="M11.8 13.2 19 5" /></svg>
          </button>
        </div>
        {!disabled ? <span className="composer-hint">Enter to send · Shift + Enter for a new line</span> : null}
      </form>

      <input ref={fileInputRef} type="file" onChange={chooseFile} hidden />
      <EmojiPicker open={pickerOpen && !disabled} triggerRef={emojiToggleRef} onSelect={onSelect} onClose={closePicker} />
      <AttachMenu open={attachMenuOpen && !disabled && !editing} triggerRef={attachToggleRef} onSelect={selectAttachOption} onClose={closeAttachMenu} />
    </div>
  )
}