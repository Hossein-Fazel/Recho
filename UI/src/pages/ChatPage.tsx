import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Composer } from '../components/Composer'
import { ConversationInfoPanel } from '../components/ConversationInfoPanel'
import { JoinGroupModal } from '../components/JoinGroupModal'
import { NewGroupModal } from '../components/NewGroupModal'
import { Sidebar } from '../components/Sidebar'
import { Thread } from '../components/Thread'
import { useAuth } from '../context/AuthContext'
import { useWebSocket } from '../hooks/useWebSocket'
import { api } from '../lib/api'
import { firstGroupMemberCursor } from '../lib/cursor'
import { newRequestId } from '../lib/id'
import { clearInvitePath } from '../lib/invite'
import type {
  Conversation,
  CreateGroupResponse,
  GroupMember,
  Message,
  UserSearch,
} from '../lib/types'

type ChatPageProps = {
  /** Invite code from a /join/<code> link the user landed on. */
  inviteCode?: string
}

export function ChatPage({ inviteCode = '' }: ChatPageProps) {
  const { user, logout } = useAuth()

  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeId, setActiveId] = useState<string | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [nextMessageCursor, setNextMessageCursor] = useState('')
  const [loadingMessages, setLoadingMessages] = useState(false)
  const [mobileChat, setMobileChat] = useState(false)
  const [notice, setNotice] = useState('')
  const [messageStatuses, setMessageStatuses] = useState<Map<number, 'sending' | 'sent'>>(new Map())
  const [editing, setEditing] = useState<{ id: number; content: string } | null>(null)
  const [pendingDelete, setPendingDelete] = useState<Message | null>(null)
  const [infoOpen, setInfoOpen] = useState(false)
  const [newGroupOpen, setNewGroupOpen] = useState(false)
  const [joinOpen, setJoinOpen] = useState(Boolean(inviteCode))
  const [joinCode, setJoinCode] = useState(inviteCode)
  const [groupSenders, setGroupSenders] = useState<Map<string, GroupMember>>(new Map())

  const loadedFor = useRef<string | null>(null)
  const pendingConvFetches = useRef<Set<string>>(new Set())

  const active = useMemo(
    () =>
      conversations.find(
        (conversation) =>
          conversation.conversation_id === activeId,
      ) ?? null,
    [conversations, activeId],
  )

  // Keep the sender details (name, username, avatar) for every group member
  // so group messages can be labeled like Telegram. Reloaded whenever the
  // active group changes.
  useEffect(() => {
    const conversationId = active?.conversation_id ?? null
    setGroupSenders(new Map())

    if (!conversationId || active?.conversation_type !== 'group') {
      return
    }

    const targetId = conversationId
    let cancelled = false

    async function loadSenders() {
      const senders = new Map<string, GroupMember>()
      let cursor = firstGroupMemberCursor

      try {
        while (cursor) {
          const result = await api.groupMembers(targetId, cursor)
          if (cancelled) return

          for (const member of result.members) {
            senders.set(member.user_id, member)
          }

          cursor = result.nextCursor
          if (!cursor || result.members.length === 0) break
        }
      } catch {
        // Sender labels are a nice-to-have; render without them on failure.
        return
      }

      if (!cancelled) setGroupSenders(senders)
    }

    void loadSenders()

    return () => {
      cancelled = true
    }
  }, [active?.conversation_id, active?.conversation_type])

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

  const applyIncomingToList = useCallback(
    (list: Conversation[], incoming: Message) => {
      const next = list.map((conversation) =>
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
    },
    [],
  )

  // A message for a conversation we don't have cached yet (e.g. the very
  // first message from someone new) can't be spliced into the sidebar with
  // just the data the WS event carries — we don't know the sender's name,
  // avatar, or whether it's a group. Fetch that conversation's info once,
  // then prepend it to the list.
  const hydrateUnknownConversation = useCallback(
    (incoming: Message) => {
      const conversationId = incoming.conversation_id

      if (pendingConvFetches.current.has(conversationId)) {
        return
      }

      pendingConvFetches.current.add(conversationId)

      api
        .conversation(conversationId)
        .then((conversation) => {
          setConversations((latest) => {
            if (
              latest.some(
                (c) => c.conversation_id === conversationId,
              )
            ) {
              return applyIncomingToList(latest, incoming)
            }

            return applyIncomingToList(
              [conversation, ...latest],
              incoming,
            )
          })
        })
        .catch(() => {
          setNotice('Could not load the new conversation')
        })
        .finally(() => {
          pendingConvFetches.current.delete(conversationId)
        })
    },
    [applyIncomingToList],
  )

  const {
    sendMessage,
    sendEditMessage,
    sendDeleteMessage,
  } = useWebSocket({
    enabled: Boolean(user),

    onMessage: (incoming) => {
      setConversations((current) => {
        const exists = current.some(
          (conversation) =>
            conversation.conversation_id ===
            incoming.conversation_id,
        )

        if (!exists) {
          hydrateUnknownConversation(incoming)
          return current
        }

        return applyIncomingToList(current, incoming)
      })

      if (incoming.conversation_id !== loadedFor.current) {
        return
      }

      setMessages((current) => {
        if (current.some((message) => message.id === incoming.id)) {
          return current
        }

        const withoutPending = incoming.request_id
          ? current.filter(
              (message) => message.request_id !== incoming.request_id,
            )
          : current

        return [
          ...withoutPending.filter(
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

    onMessageEdited: (edited) => {
      // If a message in the active thread was edited, update it in place.
      if (edited.conversation_id === loadedFor.current) {
        setMessages((current) =>
          current.map((message) =>
            message.id === edited.id
              ? {
                  ...message,
                  content: edited.content,
                  updated_at: edited.updated_at,
                  edited: true,
                }
              : message,
          ),
        )
      }

      // Update the sidebar last-message preview if needed.
      setConversations((current) =>
        current.map((conversation) =>
          conversation.conversation_id === edited.conversation_id &&
          conversation.last_message_id === edited.id
            ? { ...conversation, last_message_content: edited.content }
            : conversation,
        ),
      )
    },

    onMessageDeleted: (deleted) => {
      if (deleted.conversation_id === loadedFor.current) {
        setMessages((current) => {
          const remaining = current.filter(
            (message) => message.id !== deleted.id,
          )

          const last = remaining[remaining.length - 1]
          if (last) {
            setConversations((convs) =>
              convs.map((conversation) =>
                conversation.conversation_id === deleted.conversation_id
                  ? {
                      ...conversation,
                      last_message_id: last.id,
                      last_message_content: last.content,
                      last_message_created_at: last.created_at,
                      updated_at: last.updated_at,
                    }
                  : conversation,
              ),
            )
          }

          return remaining
        })
      } else {
        // The deleted message was in another conversation; mark the list
        // stale by reloading conversations.
        void refreshConversations()
      }
    },

    onError: (message) => {
      setNotice(message)
    },
  })

  function selectConversation(conversation: Conversation) {
    setActiveId(conversation.conversation_id)
    setMobileChat(true)
    setNotice('')
    setEditing(null)
    setInfoOpen(false)
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
    setInfoOpen(false)


  }

  function onGroupCreated(group: CreateGroupResponse) {
    setConversations((current) => {
      if (
        current.some(
          (conversation) => conversation.conversation_id === group.conversation_id,
        )
      ) {
        return current
      }

      const created: Conversation = {
        conversation_id: group.conversation_id,
        conversation_type: 'group',
        user_id: '',
        username: '',
        display_name: '',
        avatar_url: '',
        group_name: group.name,
        group_avatar_url: group.avatar_url,
        last_message_id: 0,
        last_message_content: '',
        last_message_created_at: '',
        updated_at: group.updated_at || new Date().toISOString(),
      }

      return [created, ...current]
    })

    setActiveId(group.conversation_id)
    setMobileChat(true)
    setNotice('')
    setInfoOpen(false)
  }

  function closeJoin() {
    setJoinOpen(false)
    setJoinCode('')
    // Drop /join/<code> from the URL so a refresh doesn't reopen the modal.
    clearInvitePath()
  }

  async function onGroupJoined(conversationId: string) {
    closeJoin()

    // The joined group isn't in the cached list yet; reload so it shows up
    // with its name, avatar and last message.
    try {
      await refreshConversations()
    } catch {
      setNotice('Joined, but the chat list could not be refreshed')
    }

    setActiveId(conversationId)
    setMobileChat(true)
    setInfoOpen(false)
  }

  function send(content: string) {
    if (!activeId || !user) {
      return
    }

    const requestId = newRequestId()

    // If we're editing a message, send the edit instead of a new message.
    if (editing) {
      setNotice('')

      const sent = sendEditMessage(
        activeId,
        editing.id,
        content,
        requestId,
      )

      if (!sent) {
        setNotice('Connecting to Recho…')
        return
      }

      // Optimistically update the edited message in place.
      setMessages((current) =>
        current.map((message) =>
          message.id === editing.id
            ? {
                ...message,
                content,
                updated_at: new Date().toISOString(),
                edited: true,
              }
            : message,
        ),
      )

      setConversations((current) =>
        current.map((conversation) =>
          conversation.conversation_id === activeId &&
          conversation.last_message_id === editing.id
            ? {
                ...conversation,
                last_message_content: content,
              }
            : conversation,
        ),
      )

      setEditing(null)
      return
    }

    const sent = sendMessage(activeId, content, requestId)

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
      request_id: requestId,
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

  function startEdit(message: Message) {
    if (message.id < 0 || message.sender_id !== user?.id) return
    setEditing({ id: message.id, content: message.content })
  }

  function cancelEdit() {
    setEditing(null)
  }

  function confirmDelete() {
    if (!pendingDelete || !activeId) {
      setPendingDelete(null)
      return
    }

    const requestId = newRequestId()
    const sent = sendDeleteMessage(
      activeId,
      pendingDelete.id,
      requestId,
    )

    setPendingDelete(null)

    if (!sent) {
      setNotice('Connecting to Recho…')
      return
    }

    // Optimistically remove the message.
    setMessages((current) =>
      current.filter((message) => message.id !== pendingDelete.id),
    )
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
        onNewGroup={() => setNewGroupOpen(true)}
        onJoinGroup={() => {
          setJoinCode('')
          setJoinOpen(true)
        }}
        onLogout={() => void logout()}
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
          senders={groupSenders}
          hasMore={Boolean(nextMessageCursor)}
          loading={loadingMessages}
          onLoadMore={() => void loadMore()}
          onBack={() => setMobileChat(false)}
          onEdit={startEdit}
          onDelete={setPendingDelete}
          onOpenInfo={() => setInfoOpen(true)}
        />

        <Composer
          disabled={!activeId}
          onSend={send}
          editing={editing}
          onCancelEdit={cancelEdit}
        />
      </main>

      <ConversationInfoPanel
        conversation={active}
        currentUserId={user.id}
        open={infoOpen}
        onClose={() => setInfoOpen(false)}
      />

      <NewGroupModal
        open={newGroupOpen}
        onClose={() => setNewGroupOpen(false)}
        onCreated={onGroupCreated}
      />

      <JoinGroupModal
        open={joinOpen}
        initialCode={joinCode}
        onClose={closeJoin}
        onJoined={(conversationId) => void onGroupJoined(conversationId)}
      />

      {pendingDelete ? (
        <div className="modal-overlay" role="dialog" aria-modal="true">
          <div className="modal">
            <h3>Delete message?</h3>
            <p>This message will be deleted for everyone.</p>
            <div className="modal-actions">
              <button
                type="button"
                className="ghost-btn"
                onClick={() => setPendingDelete(null)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="danger-btn"
                onClick={confirmDelete}
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>

  )
}
