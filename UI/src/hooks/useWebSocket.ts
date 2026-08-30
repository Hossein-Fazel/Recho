import { useEffect, useRef, useState } from 'react'
import type { MessageCreated, WSResponse } from '../lib/types'

type Handlers = {
  onMessage: (message: MessageCreated) => void
  enabled: boolean
}

export type WebSocketStatus =
  | 'connecting'
  | 'connected'
  | 'disconnected'

export function useWebSocket({ onMessage, enabled }: Handlers) {
  const onMessageRef = useRef(onMessage)
  const socketRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<number | null>(null)
  const shouldReconnectRef = useRef(false)

  const [status, setStatus] = useState<WebSocketStatus>(
    enabled ? 'connecting' : 'disconnected',
  )

  onMessageRef.current = onMessage

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
          const data = payload.data as MessageCreated | undefined

          if (
            (payload.type === 'message.created' || !payload.type) &&
            data?.conversation_id &&
            data.content
          ) {
            onMessageRef.current(data)
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
  ): boolean {
    const socket = socketRef.current

    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return false
    }

    socket.send(
      JSON.stringify({
        type: 'message.create',
        payload: {
          content,
          conversation_id: conversationId,
        },
      }),
    )

    return true


  }

  return {
    sendMessage,
    status,
    isConnected: status === 'connected',
  }
}
