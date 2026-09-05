import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
} from 'react'
import { Avatar } from './Avatar'
import {
  conversationTitle,
  formatMessageTime,
  hueFromId,
} from '../lib/format'
import type {
  Conversation,
  GroupMember,
  Message,
  User,
} from '../lib/types'

type ThreadProps = {
  user: User
  conversation: Conversation | null
  messages: Message[]
  messageStatuses: Map<number, 'sending' | 'sent'>
  senders?: ReadonlyMap<string, GroupMember>
  hasMore: boolean
  loading: boolean
  onLoadMore: () => void
  onBack: () => void
  onEdit: (message: Message) => void
  onDelete: (message: Message) => void
  onOpenInfo: () => void
}

type MenuState = {
  message: Message
  x: number
  y: number
} | null

export function Thread({
  user,
  conversation,
  messages,
  messageStatuses,
  senders,
  hasMore,
  loading,
  onLoadMore,
  onBack,
  onEdit,
  onDelete,
  onOpenInfo,
}: ThreadProps) {
  const scroller = useRef<HTMLDivElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  const [menu, setMenu] = useState<MenuState>(null)

  const stickToBottom = useRef(true)
  const previousConversation = useRef<string | null>(null)

  // Used to preserve the scroll position when older messages are prepended.
  const loadingOlder = useRef(false)
  const previousScrollHeight = useRef(0)

  useEffect(() => {
    const conversationId = conversation?.conversation_id ?? null

    if (conversationId !== previousConversation.current) {
      previousConversation.current = conversationId
      stickToBottom.current = true
      loadingOlder.current = false
    }

  }, [conversation?.conversation_id])

  useEffect(() => {
    const el = scroller.current

    if (!el) return

    // Older messages were prepended.
    // Keep the user's viewport at exactly the same messages.
    if (loadingOlder.current) {
      const heightDifference =
        el.scrollHeight - previousScrollHeight.current

      el.scrollTop += heightDifference

      loadingOlder.current = false
      return
    }

    // Initial conversation load or a new message while already near bottom.
    if (stickToBottom.current) {
      el.scrollTop = el.scrollHeight
    }

  }, [messages])

  // Close the context menu on outside click or Escape.
  useEffect(() => {
    if (!menu) return

    function onClick(event: MouseEvent) {
      if (menuRef.current?.contains(event.target as Node)) return
      setMenu(null)
    }

    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') setMenu(null)
    }

    document.addEventListener('mousedown', onClick)
    document.addEventListener('keydown', onKey)

    return () => {
      document.removeEventListener('mousedown', onClick)
      document.removeEventListener('keydown', onKey)
    }
  }, [menu])

  function onScroll() {
    const el = scroller.current

    if (!el) return

    const distanceFromBottom =
      el.scrollHeight - el.scrollTop - el.clientHeight

    stickToBottom.current = distanceFromBottom < 80

    if (
      el.scrollTop <= 80 &&
      hasMore &&
      !loading &&
      !loadingOlder.current
    ) {
      loadingOlder.current = true
      previousScrollHeight.current = el.scrollHeight

      onLoadMore()
    }

  }

  function openMenu(
    event: React.MouseEvent,
    message: Message,
  ) {
    event.preventDefault()
    setMenu({ message, x: event.clientX, y: event.clientY })
  }

  function copyMessage(message: Message) {
    void navigator.clipboard.writeText(message.content)
    setMenu(null)
  }

  type MessageRun = {
    messages: Message[]
    mine: boolean
    member: GroupMember | null
  }

  // Consecutive messages from the same sender are one "run". In groups a run
  // is rendered as a column with a single avatar and one name label, the way
  // Telegram does it.
  const runs = useMemo<MessageRun[]>(() => {
    const result: MessageRun[] = []

    for (const message of messages) {
      const current = result[result.length - 1]
      const currentLast = current?.messages[current.messages.length - 1]

      if (current && currentLast?.sender_id === message.sender_id) {
        current.messages.push(message)
      } else {
        result.push({
          messages: [message],
          mine: message.sender_id === user.id,
          member: senders?.get(message.sender_id) ?? null,
        })
      }
    }

    return result
  }, [messages, senders, user.id])

  function senderLabel(member: GroupMember | null): string {
    if (!member) return ''
    return member.display_name?.trim() || member.username?.trim() || ''
  }

  function renderBubble(
    message: Message,
    stacked: boolean,
  ) {
    const mine = message.sender_id === user.id

    const status = mine
      ? message.id < 0
        ? 'sending'
        : messageStatuses.get(message.id)
      : undefined

    return (
      <article
        key={message.id}
        className={`bubble ${mine ? 'mine' : ''} ${stacked ? 'stacked' : ''
          }`}
        onContextMenu={(event) => openMenu(event, message)}
      >
        <p>
          {message.content}
          {message.edited ? (
            <span className="edited-tag"> edited</span>
          ) : null}
        </p>

        <time className="bubble-time">
          {status !== 'sending' && (
            <span className="bubble-time-text">
              {formatMessageTime(message.created_at)}
            </span>
          )}

          {status ? (
            <span
              className={`msg-status ${status} ${status === 'sending' ? 'spinning' : ''
                }`}
              aria-label={
                status === 'sending'
                  ? 'Sending'
                  : 'Sent'
              }
            >
              {status === 'sending' ? (
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <circle cx="12" cy="12" r="9" />
                  <path d="M12 7v5l3 2" />
                </svg>
              ) : (
                <svg
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2.2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M4 12.5l5 5L20 6.5" />
                </svg>
              )}
            </span>
          ) : null}
        </time>
      </article>
    )
  }

  if (!conversation) {
    return (<section className="thread empty-thread"> <div className="empty-card"> <span className="brand-mark lg" aria-hidden> <span /> <span /> <span /> </span>


      <h2>Pick a conversation</h2>

      <p>
        Search a username and start a live thread. New messages
        appear instantly.
      </p>
    </div>
    </section>
    )


  }

  const title = conversationTitle(conversation)
  const isGroup = conversation.conversation_type === 'group'

  return (<section className="thread"> <header className="thread-head"> <button
    type="button"
    className="ghost back"
    onClick={onBack}
    aria-label="Back to conversations"
  >
    Chats </button>

    <button
      type="button"
      className="thread-peer"
      onClick={onOpenInfo}
      aria-label={`Open ${conversation.conversation_type === 'group' ? 'group info' : 'profile'} for ${title}`}
    >
      <Avatar
        id={conversation.user_id || conversation.conversation_id}
        name={title}
        url={
          conversation.avatar_url ||
          conversation.group_avatar_url
        }
        size="sm"
      />

      <div className="thread-title">
        <h2>{title}</h2>

        <p className="eyebrow">
          {conversation.conversation_type === 'group'
            ? 'Group'
            : 'Direct'}{' '}
          · live
        </p>
      </div>
    </button>
  </header>

    <div
      className="messages"
      ref={scroller}
      onScroll={onScroll}
    >
      {loading && loadingOlder.current ? (
        <div className="messages-loading">
          Loading earlier messages…
        </div>
      ) : null}

      {isGroup
        ? runs.map((run, runIndex) => {
          const label = senderLabel(run.member)
          const firstId = run.messages[0].id

          return (
            <div
              key={`run:${firstId}:${runIndex}`}
              className={`msg-group${run.mine ? ' mine' : ''}`}
            >
              {run.mine ? null : (
                <div className="msg-gutter">
                  {run.member && label ? (
                    <Avatar
                      id={run.member.user_id}
                      name={label}
                      url={run.member.avatar_url || undefined}
                      size="sm"
                    />
                  ) : null}
                </div>
              )}

              <div className="msg-stack">
                {!run.mine && run.member && label ? (
                  <span
                    className="msg-author"
                    title={
                      run.member.username
                        ? `@${run.member.username}`
                        : label
                    }
                    style={{
                      '--author-hue': hueFromId(run.member.user_id),
                    } as CSSProperties}
                  >
                    {label}
                  </span>
                ) : null}

                {run.messages.map((message, index) =>
                  renderBubble(message, index > 0),
                )}
              </div>
            </div>
          )
        })
        : messages.map((message, index) => {
          const previous = messages[index - 1]
          const stacked =
            Boolean(previous) &&
            previous.sender_id === message.sender_id

          return renderBubble(message, stacked)
        })}
    </div>

    {menu ? (
      <div
        ref={menuRef}
        className="context-menu"
        style={{ top: menu.y, right: window.innerWidth - menu.x }}
      >
        <button
          type="button"
          className="context-item"
          onClick={() => copyMessage(menu.message)}
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <rect x="9" y="9" width="11" height="11" rx="2" />
            <path d="M5 15V6a1 1 0 0 1 1-1h9" />
          </svg>
          Copy message
        </button>

        {menu.message.sender_id === user.id ? (
          <button
            type="button"
            className="context-item"
            onClick={() => {
              onEdit(menu.message)
              setMenu(null)
            }}
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M17 3a2.8 2.8 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
            </svg>
            Edit
          </button>
        ) : null}

        {menu.message.sender_id === user.id ? (
          <button
            type="button"
            className="context-item danger"
            onClick={() => {
              onDelete(menu.message)
              setMenu(null)
            }}
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M4 7h16" />
              <path d="M10 11v6M14 11v6" />
              <path d="M6 7l1 13h10l1-13" />
              <path d="M9 7V4h6v3" />
            </svg>
            Delete
          </button>
        ) : null}
      </div>
    ) : null}
  </section>

  )
}
