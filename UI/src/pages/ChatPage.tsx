import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Composer } from '../components/Composer'
import { Sidebar } from '../components/Sidebar'
import { Thread } from '../components/Thread'
import { useAuth } from '../context/AuthContext'
import { useWebSocket } from '../hooks/useWebSocket'
import { api } from '../lib/api'
import type { Conversation, Message, UserSearch } from '../lib/types'

export function ChatPage() {
  const { user, logout } = useAuth()

  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeId, setActiveId] = useState<string | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [nextMessageCursor, setNextMessageCursor] = useState('')
  const [loadingMessages, setLoadingMessages] = useState(false)
  const [mobileChat, setMobileChat] = useState(false)
  const [notice, setNotice] = useState('')
  const [messageStatuses, setMessageStatuses] = useState<Map<number, 'sending' | 'sent'>>(new Map())

  const loadedFor = useRef<string | null>(null)

  const active = useMemo(
    () =>
      conversations.find(
        (conversation) =>
          conversation.conversation_id === activeId,
      ) ?? null,
    [conversations, activeId],
  )

  const refreshConversations = useCallback(async () => {
    const response = await api.conversations()
    setConversations(response.conversations)
  }, [])

  useEffect(() => {
    void refreshConversations().catch(() => {
      setNotice('Could not load conversations')
    })
  }, [refreshConversations])

  const loadThread = useCallback(async (conversationId: string) => {
    setLoadingMessages(true)

    try {
      const response = await api.messages(conversationId)

      const chronological = [...response.messages].reverse()

      setMessages(chronological)
      setMessageStatuses(
        new Map(chronological.map((m) => [m.id, 'sent' as const])),
      )
      setNextMessageCursor(response.nextCursor)
      loadedFor.current = conversationId
    } finally {
      setLoadingMessages(false)
    }

  }, [])

  useEffect(() => {
    if (!activeId) {
      setMessages([])
      setMessageStatuses(new Map())
      setNextMessageCursor('')
      loadedFor.current = null
      return
    }

    void loadThread(activeId).catch(() => {
      setNotice('Could not load messages')
    })

  }, [activeId, loadThread])

  const loadMore = useCallback(async () => {
    if (!activeId || !nextMessageCursor || loadingMessages) {
      return
    }

    setLoadingMessages(true)

    try {
      const response = await api.messages(
        activeId,
        nextMessageCursor,
      )

      const older = [...response.messages].reverse()

      setMessageStatuses((statuses) => {
        const updated = new Map(statuses)
        older.forEach((m) => updated.set(m.id, 'sent'))
        return updated
      })

      setMessages((current) => [...older, ...current])
      setNextMessageCursor(response.nextCursor)
    } finally {
      setLoadingMessages(false)
    }


  }, [activeId, nextMessageCursor, loadingMessages])

  const { sendMessage } = useWebSocket({
    enabled: Boolean(user),

    onMessage: (incoming) => {
      setConversations((current) => {
        const next = current.map((conversation) =>
          conversation.conversation_id ===
            incoming.conversation_id
            ? {
              ...conversation,
              last_message_id: incoming.id,
              last_message_content: incoming.content,
              last_message_created_at:
                incoming.created_at,
              updated_at:
                incoming.updated_at ||
                incoming.created_at,
            }
            : conversation,
        )

        next.sort((a, b) => {
          const timeA = new Date(
            a.updated_at || a.last_message_created_at,
          ).getTime()

          const timeB = new Date(
            b.updated_at || b.last_message_created_at,
          ).getTime()

          return timeB - timeA
        })

        return next
      })

      if (incoming.conversation_id !== loadedFor.current) {
        return
      }

      setMessages((current) => {
        if (current.some((message) => message.id === incoming.id)) {
          return current
        }

        return [
          ...current.filter(
            (message) =>
              message.id > 0 ||
              message.content !== incoming.content,
          ),
          incoming,
        ]
      })

      setMessageStatuses((statuses) => {
        const updated = new Map(statuses)
        updated.set(incoming.id, 'sent')
        return updated
      })
    },


  })

  function selectConversation(conversation: Conversation) {
    setActiveId(conversation.conversation_id)
    setMobileChat(true)
    setNotice('')
  }

  function onCreated(
    conversationId: string,
    peer: UserSearch,
  ) {
    setConversations((current) => {
      if (
        current.some(
          (conversation) =>
            conversation.conversation_id === conversationId,
        )
      ) {
        return current
      }

      const created: Conversation = {
        conversation_id: conversationId,
        conversation_type: 'direct',
        user_id: peer.id,
        username: peer.username,
        display_name: peer.display_name,
        avatar_url: peer.avatar_url,
        group_name: '',
        group_avatar_url: '',
        last_message_id: 0,
        last_message_content: '',
        last_message_created_at: '',
        updated_at: new Date().toISOString(),
      }

      return [created, ...current]
    })

    setActiveId(conversationId)
    setMobileChat(true)
    setNotice('')


  }

  function send(content: string) {
    if (!activeId || !user) {
      return
    }

    const sent = sendMessage(activeId, content)

    if (!sent) {
      setNotice('Connecting to Recho…')
      return
    }

    const now = new Date().toISOString()

    const optimisticMessage: Message = {
      id: -Date.now(),
      conversation_id: activeId,
      sender_id: user.id,
      content,
      created_at: now,
      updated_at: now,
    }

    setMessages((current) => [
      ...current,
      optimisticMessage,
    ])

    setMessageStatuses((statuses) => {
      const updated = new Map(statuses)
      updated.set(optimisticMessage.id, 'sending')
      return updated
    })

    setConversations((current) =>
      current.map((conversation) =>
        conversation.conversation_id === activeId
          ? {
            ...conversation,
            last_message_content: content,
            updated_at: now,
          }
          : conversation,
      ),
    )

    setNotice('')

  }

  if (!user) {
    return null
  }

  return (
    <div
      className={`app-shell ${mobileChat ? 'show-thread' : ''
        }`}
    >
      <Sidebar
        user={user}
        conversations={conversations}
        activeId={activeId}
        onSelect={selectConversation}
        onCreated={onCreated}
        onLogout={() => void logout()}
        onCloseMobile={() => setMobileChat(false)}
      />

      <main className="main">
        {notice ? (
          <p className="toast" role="status">
            {notice}
          </p>
        ) : null}

        <Thread
          user={user}
          conversation={active}
          messages={messages}
          messageStatuses={messageStatuses}
          hasMore={Boolean(nextMessageCursor)}
          loading={loadingMessages}
          onLoadMore={() => void loadMore()}
          onBack={() => setMobileChat(false)}
        />

        <Composer
          disabled={!activeId}
          onSend={send}
        />
      </main>
    </div>

  )
}
