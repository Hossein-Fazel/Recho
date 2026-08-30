import { useEffect, useRef } from 'react'
import { Avatar } from './Avatar'
import { displayName, formatMessageTime } from '../lib/format'
import type { Conversation, Message, User } from '../lib/types'

type ThreadProps = {
  user: User
  conversation: Conversation | null
  messages: Message[]
  hasMore: boolean
  loading: boolean
  onLoadMore: () => void
  onBack: () => void
}

export function Thread({
  user,
  conversation,
  messages,
  hasMore,
  loading,
  onLoadMore,
  onBack,
}: ThreadProps) {
  const scroller = useRef<HTMLDivElement>(null)

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

  const title = displayName(conversation)

  return (<section className="thread"> <header className="thread-head"> <button
    type="button"
    className="ghost back"
    onClick={onBack}
    aria-label="Back to conversations"
  >
    Chats </button>

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

      {messages.map((message, index) => {
        const mine = message.sender_id === user.id
        const previous = messages[index - 1]

        const stacked =
          Boolean(previous) &&
          previous.sender_id === message.sender_id

        return (
          <article
            key={message.id}
            className={`bubble ${mine ? 'mine' : ''} ${stacked ? 'stacked' : ''
              }`}
          >
            <p>{message.content}</p>

            <time>
              {formatMessageTime(message.created_at)}
            </time>
          </article>
        )
      })}
    </div>
  </section>

  )
}
