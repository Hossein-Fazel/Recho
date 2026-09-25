import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Composer } from '../components/Composer'
import { ConversationInfoPanel } from '../components/ConversationInfoPanel'
import { UserInfoPanel } from '../components/UserInfoPanel'
import { JoinGroupModal } from '../components/JoinGroupModal'
import { NewGroupModal } from '../components/NewGroupModal'
import { ProfileModal } from '../components/ProfileModal'
import { Sidebar } from '../components/Sidebar'
import { Thread } from '../components/Thread'
import { useAuth } from '../context/AuthContext'
import { useWebSocket } from '../hooks/useWebSocket'
import { api } from '../lib/api'
import { firstGroupMemberCursor } from '../lib/cursor'
import { messagePreview } from '../lib/format'
import { newRequestId } from '../lib/id'
import { clearInvitePath } from '../lib/invite'
import type {
  Conversation,
  FileCategory,
  CreateGroupResponse,
  GroupMember,
  Message,
  MessageFile,
  UpdateGroupResponse,
  UserSearch,
  PresenceUpdated,
} from '../lib/types'

type ChatPageProps = {
  /** Invite code from a /join/<code> link the user landed on. */
  inviteCode?: string
}

export function ChatPage({ inviteCode = '' }: ChatPageProps) {
  const { user, logout } = useAuth()

  const [conversations, setConversations] = useState<Conversation[]>([])
  const [onlineUserIds, setOnlineUserIds] = useState<Set<string>>(new Set())
  const [activeId, setActiveId] = useState<string | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [nextMessageCursor, setNextMessageCursor] = useState('')
  const [loadingMessages, setLoadingMessages] = useState(false)
  const [mobileChat, setMobileChat] = useState(false)
  const [notice, setNotice] = useState('')
  const [messageStatuses, setMessageStatuses] = useState<Map<number, 'sending' | 'sent'>>(new Map())
  const [editing, setEditing] = useState<{ id: number; content: string; file: MessageFile | null } | null>(null)
  const [pendingDelete, setPendingDelete] = useState<Message | null>(null)
  const [infoOpen, setInfoOpen] = useState(false)
  const [memberProfileId, setMemberProfileId] = useState<string | null>(null)
  const [newGroupOpen, setNewGroupOpen] = useState(false)
  const [profileOpen, setProfileOpen] = useState(false)
  const [joinOpen, setJoinOpen] = useState(Boolean(inviteCode))
  const [joinCode, setJoinCode] = useState(inviteCode)
  const [groupSenders, setGroupSenders] = useState<Map<string, GroupMember>>(new Map())
  const [groupInfoVersion, setGroupInfoVersion] = useState(0)
  const [uploadStates, setUploadStates] = useState<Map<string, { conversationId: string; message: Message; progress: number; status: 'uploading' | 'sending'; caption: string; socketSent: boolean }>>(new Map())
  const [selectedFiles, setSelectedFiles] = useState<Map<string, { conversationId: string; file: File; category: FileCategory; caption: string }>>(new Map())

  const uploadControllers = useRef<Map<string, AbortController>>(new Map())
  const uploadPreviews = useRef<Map<string, string>>(new Map())
  const optimisticId = useRef(-1)
  const uploadStatesRef = useRef(uploadStates)
  uploadStatesRef.current = uploadStates
  const selectedFilesRef = useRef(selectedFiles)
  selectedFilesRef.current = selectedFiles

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

  const directUserIds = useMemo(
    () =>
      Array.from(
        new Set(
          conversations
            .filter((conversation) => conversation.conversation_type === 'direct' && conversation.user_id)
            .map((conversation) => conversation.user_id),
        ),
      ),
    [conversations],
  )

  const directUserIdsKey = directUserIds.join(',')

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

  // The conversation summary contains the peer's last_seen. Fetch it when a
  // direct chat opens so the header uses the freshest server value instead of
  // relying on the sidebar cache.
  useEffect(() => {
    if (!activeId || active?.conversation_type !== 'direct') return

    let cancelled = false

    api.conversation(activeId)
      .then((conversation) => {
        if (cancelled) return
        setConversations((current) =>
          current.map((item) =>
            item.conversation_id === activeId
              ? { ...item, last_seen: conversation.last_seen }
              : item,
          ),
        )
      })
      .catch(() => {
        // The message thread can still be used if the summary refresh fails.
      })

    return () => {
      cancelled = true
    }
  }, [activeId, active?.conversation_type])

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
            last_message_type: incoming.type,
            last_file_category:
              incoming.type === 'file'
                ? incoming.file?.category
                : undefined,
            last_message_text: messagePreview(incoming),
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
    sendFileMessage,
    sendEditMessage,
    sendEditFileMessage,
    sendDeleteMessage,
    status: webSocketStatus,
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

      const requestId = incoming.request_id
      if (requestId) {
        const pending = uploadStatesRef.current.get(requestId)
        if (pending) {
          const preview = uploadPreviews.current.get(requestId)
          if (preview) URL.revokeObjectURL(preview)
          uploadPreviews.current.delete(requestId)
          setUploadStates((states) => {
            const next = new Map(states)
            next.delete(requestId)
            return next
          })
        }
      }

      if (incoming.conversation_id !== loadedFor.current) {
        return
      }

      setMessages((current) => {
        if (current.some((message) => message.id === incoming.id)) {
          return current
        }

        const withoutPending = incoming.request_id
          ? current.filter((message) => message.request_id !== incoming.request_id)
          : current

        return [...withoutPending, incoming]
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
                type: edited.type,
                text: edited.type === 'text' ? edited.text : null,
                file: edited.type === 'file' ? edited.file : null,
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
            ? {
              ...conversation,
              last_message_type: edited.type,
              last_file_category:
                edited.type === 'file'
                  ? edited.file?.category
                  : undefined,
              last_message_text: messagePreview(edited),
            }
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
                    last_message_type: last.type,
                    last_file_category:
                      last.type === 'file'
                        ? last.file?.category
                        : undefined,
                    last_message_text: messagePreview(last),
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

    onConversationDeleted: (deleted) => {
      setConversations((current) =>
        current.filter(
          (conversation) =>
            conversation.conversation_id !== deleted.conversation_id,
        ),
      )

      if (deleted.conversation_id === loadedFor.current) {
        setActiveId((current) =>
          current === deleted.conversation_id ? null : current,
        )
        setMessages([])
        setMessageStatuses(new Map())
        setNextMessageCursor('')
        setInfoOpen(false)
        setMobileChat(false)
      }
    },

    onGroupUpdated: (updated) => {
      setConversations((current) =>
        current.map((conversation) =>
          conversation.conversation_id === updated.group_id
            ? {
              ...conversation,
              group_name: updated.name,
              group_avatar_url: updated.avatar_url,
            }
            : conversation,
        ),
      )

      if (updated.group_id === activeId) {
        setGroupInfoVersion((version) => version + 1)
      }
    },

    onPresenceUpdated: (presence: PresenceUpdated) => {
      setOnlineUserIds((current) => {
        const next = new Set(current)

        if (presence.online) {
          next.add(presence.user_id)
        } else {
          next.delete(presence.user_id)
        }

        return next
      })

      if (!presence.online && presence.last_seen) {
        setConversations((current) =>
          current.map((conversation) =>
            conversation.user_id === presence.user_id
              ? { ...conversation, last_seen: presence.last_seen! }
              : conversation,
          ),
        )
      }
    },
  })

  // Subscribe to every direct-chat peer. Subscriptions are server-side, so
  // repeat this after every WebSocket reconnect: the backend removes the
  // subscriber's subscriptions when its socket disconnects.
  useEffect(() => {
    if (!user || directUserIds.length === 0) {
      setOnlineUserIds(new Set())
      return
    }

    let cancelled = false
    const ids = directUserIdsKey.split(',').filter(Boolean)

    async function syncPresence() {
      try {
        if (webSocketStatus === 'connected') {
          await api.subscribePresence(ids)
        }

        const response = await api.presence(ids)

        if (cancelled) return

        setOnlineUserIds(
          new Set(
            response.statuses
              .filter((status) => status.online)
              .map((status) => status.user_id),
          ),
        )
      } catch {
        // Presence is supplemental; chat remains usable if this request fails.
      }
    }

    void syncPresence()

    return () => {
      cancelled = true
      if (webSocketStatus === 'connected') {
        void api.unsubscribePresence(ids)
      }
    }
  }, [user, directUserIdsKey, webSocketStatus])

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
        last_message_text: '',
        last_message_created_at: '',
        updated_at: new Date().toISOString(),
        last_seen: '',
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
        last_message_text: '',
        last_message_created_at: '',
        updated_at: group.updated_at || new Date().toISOString(),
        last_seen: '',
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

  function onConversationLeft(conversationId: string, deleted: boolean) {
    setConversations((current) =>
      current.filter(
        (conversation) => conversation.conversation_id !== conversationId,
      ),
    )

    setActiveId((current) => (current === conversationId ? null : current))
    setMessages([])
    setMessageStatuses(new Map())
    setNextMessageCursor('')
    setInfoOpen(false)
    setMobileChat(false)
    setNotice(deleted ? 'Group deleted' : 'You left the group')
  }

  function onGroupInfoUpdated(group: UpdateGroupResponse) {
    setConversations((current) =>
      current.map((conversation) =>
        conversation.conversation_id === group.id
          ? {
            ...conversation,
            group_name: group.name,
            group_avatar_url: group.avatar_url,
          }
          : conversation,
      ),
    )
  }

  const pendingUploadMessages = useMemo(() => {
    const pending: Message[] = []
    for (const upload of uploadStates.values()) {
      if (upload.conversationId === activeId) pending.push(upload.message)
    }
    return pending
  }, [uploadStates, activeId])

  const displayedMessages = useMemo(() => {
    const merged = [...messages, ...pendingUploadMessages]
    return merged.sort((a, b) => {
      const byTime = new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      if (byTime !== 0) return byTime
      return a.id - b.id
    })
  }, [messages, pendingUploadMessages])

  function removeUpload(clientMessageId: string) {
    const controller = uploadControllers.current.get(clientMessageId)
    controller?.abort()
    uploadControllers.current.delete(clientMessageId)

    const preview = uploadPreviews.current.get(clientMessageId)
    if (preview) URL.revokeObjectURL(preview)
    uploadPreviews.current.delete(clientMessageId)

    setUploadStates((current) => {
      const next = new Map(current)
      next.delete(clientMessageId)
      return next
    })
  }

  function updateUpload(clientMessageId: string, updater: (current: { conversationId: string; message: Message; progress: number; status: 'uploading' | 'sending'; caption: string; socketSent: boolean }) => { conversationId: string; message: Message; progress: number; status: 'uploading' | 'sending'; caption: string; socketSent: boolean }) {
    setUploadStates((current) => {
      const upload = current.get(clientMessageId)
      if (!upload) return current
      const next = new Map(current)
      next.set(clientMessageId, updater(upload))
      return next
    })
  }

  function onFileSelected(file: File, category: FileCategory): string | null {
    if (!activeId || !user) return null

    const requestId = newRequestId()
    setSelectedFiles((current) => {
      const next = new Map(current)
      next.set(requestId, { conversationId: activeId, file, category, caption: '' })
      return next
    })

    return requestId
  }

  function onFileCaptionChange(clientMessageId: string, caption: string) {
    setSelectedFiles((current) => {
      const selected = current.get(clientMessageId)
      if (!selected) return current
      const next = new Map(current)
      next.set(clientMessageId, { ...selected, caption })
      return next
    })

    updateUpload(clientMessageId, (current) => ({
      ...current,
      caption,
      message: {
        ...current.message,
        file: current.message.file ? { ...current.message.file, caption } : current.message.file,
      },
    }))
  }

  function onFileSend(clientMessageId: string) {
    const selected = selectedFilesRef.current.get(clientMessageId)
    if (!selected || !activeId || !user) return

    setSelectedFiles((current) => {
      const next = new Map(current)
      next.delete(clientMessageId)
      return next
    })

    const { file, category, caption } = selected
    const requestId = clientMessageId
    const previewUrl = URL.createObjectURL(file)
    uploadPreviews.current.set(requestId, previewUrl)

    const now = new Date().toISOString()
    const optimisticMessage: Message = {
      id: optimisticId.current--,
      conversation_id: activeId,
      sender_id: user.id,
      type: 'file',
      file: {
        key: '',
        url: category === 'image' || category === 'video' ? previewUrl : undefined,
        category,
        content_type: file.type,
        size: file.size,
        file_name: file.name,
        caption,
      },
      created_at: now,
      updated_at: now,
      request_id: requestId,
      upload_status: 'uploading',
      upload_progress: 0,
      local_preview_url: previewUrl,
    }

    setUploadStates((current) => {
      const next = new Map(current)
      next.set(requestId, {
        conversationId: activeId,
        message: optimisticMessage,
        progress: 0,
        status: 'uploading',
        caption,
        socketSent: false,
      })
      return next
    })

    const controller = new AbortController()
    uploadControllers.current.set(requestId, controller)

    void api.uploadMessageMedia(activeId, category, file, {
      signal: controller.signal,
      onProgress: (progress) => updateUpload(requestId, (current) => ({
        ...current,
        progress,
        message: { ...current.message, upload_progress: progress },
      })),
    }).then((uploaded) => {
      const current = uploadStatesRef.current.get(requestId)
      if (!current) return

      updateUpload(requestId, (state) => ({
        ...state,
        status: 'sending',
        progress: 100,
        message: {
          ...state.message,
          file: { ...uploaded, caption: state.caption },
          upload_status: 'sending',
          upload_progress: 100,
        },
      }))

      const sent = sendFileMessage(current.conversationId, uploaded, current.caption, requestId)
      updateUpload(requestId, (state) => ({ ...state, socketSent: sent }))
      if (!sent) setNotice('Connecting to Recho…')
    }).catch((error) => {
      if (error instanceof DOMException && error.name === 'AbortError') return
      setNotice(error instanceof Error ? error.message : 'Upload failed')
      removeUpload(requestId)
    }).finally(() => {
      uploadControllers.current.delete(requestId)
    })

  }

  useEffect(() => {
    if (webSocketStatus !== 'connected') return

    for (const [requestId, upload] of uploadStatesRef.current) {
      if (upload.status !== 'sending' || upload.socketSent || !upload.message.file?.key) continue
      const sent = sendFileMessage(
        upload.conversationId,
        upload.message.file,
        upload.caption,
        requestId,
      )
      if (sent) {
        updateUpload(requestId, (state) => ({ ...state, socketSent: true }))
      }
    }
  }, [webSocketStatus])

  function send(content: string) {
    if (!activeId || !user) {
      return
    }

    const requestId = newRequestId()

    // If we're editing a message, send the edit instead of a new message.
    if (editing) {
      setNotice('')

      const editingFile = editing.file

      const sent = editingFile
        ? sendEditFileMessage(
          activeId,
          editing.id,
          editingFile,
          content,
          requestId,
        )
        : sendEditMessage(
          activeId,
          editing.id,
          content,
          requestId,
        )

      if (!sent) {
        setNotice('Connecting to Recho…')
        return
      }

      const editedAt = new Date().toISOString()

      // Optimistically update the edited message in place.
      setMessages((current) =>
        current.map((message) =>
          message.id === editing.id
            ? editingFile
              ? {
                ...message,
                type: 'file',
                file: { ...editingFile, caption: content },
                updated_at: editedAt,
                edited: true,
              }
              : {
                ...message,
                type: 'text',
                text: { content },
                updated_at: editedAt,
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
              last_message_type: editingFile ? 'file' : 'text',
              last_file_category: editingFile?.category,
              last_message_text: content,
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
      type: 'text',
      text: { content },
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
            last_message_type: 'text',
            last_file_category: undefined,
            last_message_text: content,
            updated_at: now,
          }
          : conversation,
      ),
    )

    setNotice('')

  }

  function startEdit(message: Message) {
    if (message.id < 0 || message.sender_id !== user?.id) return

    if (message.type === 'file') {
      setEditing({
        id: message.id,
        content: message.file?.caption ?? '',
        file: message.file ?? null,
      })
      return
    }

    setEditing({
      id: message.id,
      content: messagePreview(message),
      file: null,
    })
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
        onOpenProfile={() => setProfileOpen(true)}
        onLogout={() => void logout()}
        onlineUserIds={onlineUserIds}
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
          messages={displayedMessages}
          messageStatuses={messageStatuses}
          senders={groupSenders}
          hasMore={Boolean(nextMessageCursor)}
          loading={loadingMessages}
          onLoadMore={() => void loadMore()}
          onBack={() => setMobileChat(false)}
          onEdit={startEdit}
          onDelete={setPendingDelete}
          onOpenInfo={() => setInfoOpen(true)}
          onOpenInvite={(code) => {
            setJoinCode(code)
            setJoinOpen(true)
          }}
          onCancelUpload={removeUpload}
          isPeerOnline={Boolean(active && active.conversation_type === 'direct' && onlineUserIds.has(active.user_id))}
        />

        <Composer
          disabled={!activeId}
          conversationId={activeId}
          onSend={send}
          onFileSelected={onFileSelected}
          onFileSend={onFileSend}
          onFileCaptionChange={onFileCaptionChange}
          selectedFile={Array.from(selectedFiles.values()).find((file) => file.conversationId === activeId) ?? null}
          fileSelected={Array.from(selectedFiles.values()).some((file) => file.conversationId === activeId)}
          editing={editing}
          onCancelEdit={cancelEdit}
        />
      </main>

      <ConversationInfoPanel
        conversation={active}
        currentUserId={user.id}
        open={infoOpen && !memberProfileId}
        onClose={() => setInfoOpen(false)}
        onMemberSelect={(userId) => {
          setInfoOpen(false)
          setMemberProfileId(userId)
        }}
        onLeft={onConversationLeft}
        onGroupUpdated={onGroupInfoUpdated}
        refreshKey={groupInfoVersion}
      />

      <UserInfoPanel
        userId={memberProfileId}
        open={Boolean(memberProfileId)}
        onClose={() => {
          setMemberProfileId(null)
          setInfoOpen(true)
        }}
      />

      <NewGroupModal
        open={newGroupOpen}
        onClose={() => setNewGroupOpen(false)}
        onCreated={onGroupCreated}
      />

      {profileOpen ? (
        <ProfileModal onClose={() => setProfileOpen(false)} />
      ) : null}

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
