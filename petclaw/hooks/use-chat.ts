"use client"

import { useState, useEffect, useCallback, useRef } from 'react'
import { getWebSocketInstance, type ChatMessage, type WSEvent } from '@/lib/api'
import { ensureBackendReadyForChat } from '@/lib/api/bootstrap'

export interface UseChatOptions {
  onMessage?: (message: ChatMessage) => void
  onError?: (error: string) => void
  onConnectionChange?: (connected: boolean) => void
}

export function useChat(options: UseChatOptions = {}) {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [isTyping, setIsTyping] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const wsRef = useRef(getWebSocketInstance())
  const optionsRef = useRef(options)

  const connectWithBootstrap = useCallback(async () => {
    setError(null)
    setIsTyping(false)

    const bootstrap = await ensureBackendReadyForChat()
    if (!bootstrap.ok) {
      const reason = bootstrap.reason || 'Backend not ready for chat'
      setIsConnected(false)
      setError(reason)
      optionsRef.current.onConnectionChange?.(false)
      optionsRef.current.onError?.(reason)
      return
    }

    wsRef.current.resetReconnectAttempts()

    try {
      await wsRef.current.connect()
    } catch (err) {
      setError('连接失败')
      console.error('WebSocket connection failed:', err)
    }
  }, [])

  useEffect(() => {
    optionsRef.current = options
  }, [options])

  useEffect(() => {
    const ws = wsRef.current

    const handleEvent = (event: WSEvent) => {
      switch (event.type) {
        case 'connected':
          setIsConnected(true)
          setError(null)
          optionsRef.current.onConnectionChange?.(true)
          break

        case 'disconnected':
          setIsConnected(false)
          setIsTyping(false)
          optionsRef.current.onConnectionChange?.(false)
          break

        case 'message':
          if (typeof event.data === 'object' && event.data) {
            const message = event.data as ChatMessage
            setMessages((prev) => {
              const existingIndex = prev.findIndex((m) => m.id === message.id)
              if (existingIndex >= 0) {
                const newMessages = [...prev]
                newMessages[existingIndex] = message
                return newMessages
              }
              return [...prev, message]
            })
            setIsTyping(message.streaming ?? false)
            optionsRef.current.onMessage?.(message)
          }
          break

        case 'typing':
          const isTypingValue = typeof event.data === 'string' ? event.data === 'true' : true
          setIsTyping(isTypingValue)
          break

        case 'error':
          const errorMsg = typeof event.data === 'string' ? event.data : 'Unknown error'
          setError(errorMsg)
          setIsTyping(false)
          optionsRef.current.onError?.(errorMsg)
          break

        case 'reconnecting':
          setError('正在重新连接...')
          setIsTyping(false)
          break

        case 'emotion_change':
          break

        case 'action_trigger':
          break
      }
    }

    const unsubscribe = ws.subscribe(handleEvent)

    connectWithBootstrap().catch((err) => {
      const message = err instanceof Error ? err.message : 'Backend bootstrap failed'
      setError(message)
      console.error('Chat bootstrap failed:', err)
    })

    return () => {
      unsubscribe()
    }
  }, [connectWithBootstrap])

  const sendMessage = useCallback((content: string) => {
    if (!content.trim()) return

    if (!wsRef.current.isConnected) {
      setError('当前连接不可用，请先重连后再发送')
      return
    }

    const userMessage: ChatMessage = {
      id: `user-${Date.now()}`,
      role: 'user',
      content: content.trim(),
      timestamp: Date.now(),
    }

    setMessages((prev) => [...prev, userMessage])
    setIsTyping(true)
    setError(null)

    wsRef.current.send(content.trim())
  }, [])

  const clearMessages = useCallback(() => {
    setMessages([])
  }, [])

  const reconnect = useCallback(() => {
    setError(null)
    setIsTyping(false)
    wsRef.current.disconnect()

    connectWithBootstrap().catch((err) => {
      setError('重新连接失败')
      console.error('Reconnection failed:', err)
    })
  }, [connectWithBootstrap])

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  return {
    messages,
    isConnected,
    isTyping,
    error,
    sendMessage,
    clearMessages,
    reconnect,
    clearError,
  }
}
