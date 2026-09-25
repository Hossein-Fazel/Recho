import { useEffect, useRef, useState } from 'react'
import type {
  ConversationDeleted,
  GroupUpdated,
  MessageCreated,
  MessageFile,
  PresenceUpdated,
  MessageDeleted,
  MessageEdited,
  WSResponse,
} from '../lib/types'

type Handlers = {
  onMessage: (message: MessageCreated) => void
  onMessageEdited: (message: MessageEdited) => void
  onMessageDeleted: (deleted: MessageDeleted) => void
  onConversationDeleted: (deleted: ConversationDeleted) => void
  onGroupUpdated: (group: GroupUpdated) => void
  onPresenceUpdated: (presence: PresenceUpdated) => void
  enabled: boolean
}

export type WebSocketStatus =
  | 'connecting'
  | 'connected'
  | 'disconnected'

/** Wire shape of the file payload expected by `message.create`/`message.edit`. */
function filePayload(file: MessageFile, caption: string) {
  return {
    key: file.key,
    category: file.category,
    content_type: file.content_type,
    size: file.size,
    file_name: file.file_name ?? '',
    caption,
  }
}

export function useWebSocket({
  onMessage,
  onMessageEdited,
  onMessageDeleted,
  onConversationDeleted,
  onGroupUpdated,
  onPresenceUpdated,
  enabled,
}: Handlers) {
  const onMessageRef = useRef(onMessage)
  const onMessageEditedRef = useRef(onMessageEdited)
  const onMessageDeletedRef = useRef(onMessageDeleted)
  const onConversationDeletedRef = useRef(onConversationDeleted)
  const onGroupUpdatedRef = useRef(onGroupUpdated)
  const onPresenceUpdatedRef = useRef(onPresenceUpdated)
  const socketRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<number | null>(null)
  const shouldReconnectRef = useRef(false)

  const [status, setStatus] = useState<WebSocketStatus>(
    enabled ? 'connecting' : 'disconnected',
  )

  onMessageRef.current = onMessage
  onMessageEditedRef.current = onMessageEdited
  onMessageDeletedRef.current = onMessageDeleted
  onConversationDeletedRef.current = onConversationDeleted
  onGroupUpdatedRef.current = onGroupUpdated
  onPresenceUpdatedRef.current = onPresenceUpdated

  useEffect(() => {
    if (!enabled) {
      shouldReconnectRef.current = false
      setStatus('disconnected')
      return
    }

    shouldReconnectRef.current = true
    setStatus('connecting')

    const protocol =
      window.location.protocol === 'https:' ? 'wss:' : 'ws:'

    const url = `${protocol}//${window.location.host}/ws/`

    function clearReconnectTimer() {
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }
    }

    function connect() {
      if (!shouldReconnectRef.current) return

      const current = socketRef.current

      if (
        current &&
        (current.readyState === WebSocket.OPEN ||
          current.readyState === WebSocket.CONNECTING)
      ) {
        return
      }

      setStatus('connecting')

      const socket = new WebSocket(url)
      socketRef.current = socket

      socket.onopen = () => {
        if (socketRef.current !== socket) return

        clearReconnectTimer()
        setStatus('connected')
      }

      socket.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data as string) as WSResponse
          const type = payload.type
          const data = payload.data as Record<string, unknown>

          if (type === 'message.create' && data) {
            const message = data as MessageCreated
            if (
              message.conversation_id &&
              (message.text?.content || message.file)
            ) {
              onMessageRef.current({
                ...message,
                request_id: payload.request_id,
              })
            }
          } else if (type === 'message.edit' && data) {
            const message = data as MessageEdited
            if (
              message.conversation_id &&
              (message.text?.content || message.file)
            ) {
              onMessageEditedRef.current({
                ...message,
                request_id: payload.request_id,
              })
            }
          } else if (type === 'message.delete' && data) {
            const deleted = data as MessageDeleted
            if (deleted.id && deleted.conversation_id) {
              onMessageDeletedRef.current({
                id: deleted.id,
                conversation_id: deleted.conversation_id,
                request_id: payload.request_id,
              })
            }
          } else if (type === 'conversation.delete' && data) {
            const deleted = data as ConversationDeleted
            if (deleted.conversation_id) {
              onConversationDeletedRef.current({
                conversation_id: deleted.conversation_id,
              })
            }
          } else if (type === 'group.update' && data) {
            const updated = data as GroupUpdated
            if (updated.group_id) {
              onGroupUpdatedRef.current({
                group_id: updated.group_id,
                name: updated.name,
                avatar_url: updated.avatar_url,
                bio: updated.bio,
              })
            }
          } else if (
            (type === 'user.online' || type === 'user.offline') &&
            data
          ) {
            const presence = data as PresenceUpdated
            if (presence.user_id) {
              onPresenceUpdatedRef.current({
                user_id: presence.user_id,
                online: type === 'user.online',
                last_seen: presence.last_seen,
              })
            }
          }
        } catch {
          // Ignore malformed WebSocket frames.
        }
      }

      socket.onerror = () => {
        socket.close()
      }

      socket.onclose = () => {
        if (socketRef.current === socket) {
          socketRef.current = null
        }

        if (!shouldReconnectRef.current) {
          setStatus('disconnected')
          return
        }

        setStatus('disconnected')
        clearReconnectTimer()

        reconnectTimerRef.current = window.setTimeout(() => {
          reconnectTimerRef.current = null
          connect()
        }, 2000)
      }
    }

    connect()

    return () => {
      shouldReconnectRef.current = false
      clearReconnectTimer()

      const socket = socketRef.current
      socketRef.current = null

      if (socket) {
        socket.onclose = null
        socket.close()
      }

      setStatus('disconnected')
    }
  }, [enabled])

  function sendMessage(
    conversationId: string,
    content: string,
    requestId: string,
  ): boolean {
    return send(
      {
        type: 'message.create',
        request_id: requestId,
        payload: {
          type: 'text',
          text: { content },
          conversation_id: conversationId,
        },
      },
    )
  }

  /**
   * `file` must be the payload returned by the upload endpoint. The backend
   * derives the public URL from `key` and ignores any URL sent here.
   */
  function sendFileMessage(
    conversationId: string,
    file: MessageFile,
    caption: string,
    requestId: string,
  ): boolean {
    return send(
      {
        type: 'message.create',
        request_id: requestId,
        payload: {
          type: 'file',
          conversation_id: conversationId,
          file: filePayload(file, caption),
        },
      },
    )
  }

  function sendEditMessage(
    conversationId: string,
    messageId: number,
    content: string,
    requestId: string,
  ): boolean {
    return send(
      {
        type: 'message.edit',
        request_id: requestId,
        payload: {
          message_id: messageId,
          conversation_id: conversationId,
          type: 'text',
          text: { content },
        },
      },
    )
  }

  /** Edits a file message's caption; the file itself is immutable. */
  function sendEditFileMessage(
    conversationId: string,
    messageId: number,
    file: MessageFile,
    caption: string,
    requestId: string,
  ): boolean {
    return send(
      {
        type: 'message.edit',
        request_id: requestId,
        payload: {
          message_id: messageId,
          conversation_id: conversationId,
          type: 'file',
          file: filePayload(file, caption),
        },
      },
    )
  }

  function sendDeleteMessage(
    conversationId: string,
    messageId: number,
    requestId: string,
  ): boolean {
    return send(
      {
        type: 'message.delete',
        request_id: requestId,
        payload: {
          message_id: messageId,
          conversation_id: conversationId,
        },
      },
    )
  }

  function send(payload: unknown): boolean {
    const socket = socketRef.current

    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return false
    }

    socket.send(JSON.stringify(payload))
    return true
  }

  return {
    sendMessage,
    sendFileMessage,
    sendEditMessage,
    sendEditFileMessage,
    sendDeleteMessage,
    status,
    isConnected: status === 'connected',
  }
}
