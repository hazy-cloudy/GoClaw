"use client"

import { useState, useEffect, useCallback, useRef } from 'react'
import { getWebSocketInstance, type ChatMessage, type WSEvent } from '@/lib/api'
import { ApiError } from '@/lib/api/client'
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
  const replyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const toErrorMessage = (err: unknown, fallback: string): string => {
    if (err instanceof ApiError) {
      return `${fallback}（HTTP ${err.status}: ${err.message}）`
    }
    if (err instanceof Error && err.message) {
      return `${fallback}（${err.message}）`
    }
    return fallback
  }

  const ensureGatewayRunning = useCallback(async (): Promise<boolean> => {
    try {
      const ready = await ensureBackendReadyForChat()
      if (ready.ok) {
        return true
      }
      wsRef.current.disconnect()
      wsRef.current.resetReconnectAttempts()
      setError(`后端未就绪：${ready.reason || '请先完成 API Key 配置'}`)
      return false
    } catch (err) {
      wsRef.current.disconnect()
      wsRef.current.resetReconnectAttempts()
      setError(toErrorMessage(err, '无法连接后端服务，请确认 launcher 已启动'))
      return false
    }
  }, [])

  const clearReplyTimeout = useCallback(() => {
    if (replyTimeoutRef.current) {
      clearTimeout(replyTimeoutRef.current)
      replyTimeoutRef.current = null
    }
  }, [])

  const startReplyTimeout = useCallback(() => {
    clearReplyTimeout()
    replyTimeoutRef.current = setTimeout(() => {
      setIsTyping(false)
      setError('等待回复超时，请检查模型配置或网络连接后重试')
    }, 45000)
  }, [clearReplyTimeout])

  // 更新 options ref
  useEffect(() => {
    optionsRef.current = options
  }, [options])

  // WebSocket 事件处理
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
          clearReplyTimeout()
          optionsRef.current.onConnectionChange?.(false)
          break

        case 'message':
          if (typeof event.data === 'object' && event.data) {
            const message = event.data as ChatMessage
            setMessages((prev) => {
              // 检查是否是流式更新（相同 id）
              const existingIndex = prev.findIndex((m) => m.id === message.id)
              if (existingIndex >= 0) {
                const newMessages = [...prev]
                newMessages[existingIndex] = message
                return newMessages
              }
              return [...prev, message]
            })
            setIsTyping(message.streaming ?? false)
            if (!message.streaming) {
              clearReplyTimeout()
            }
            optionsRef.current.onMessage?.(message)
          }
          break

        case 'typing':
          const isTypingValue = typeof event.data === 'string' ? event.data === 'true' : true
          setIsTyping(isTypingValue)
          if (!isTypingValue) {
            clearReplyTimeout()
          }
          break

        case 'error':
          const errorMsg = typeof event.data === 'string' ? event.data : 'Unknown error'
          setError(errorMsg)
          setIsTyping(false)
          clearReplyTimeout()
          optionsRef.current.onError?.(errorMsg)
          break

        case 'reconnecting':
          setError('正在重新连接...')
          setIsTyping(false)
          clearReplyTimeout()
          break
      }
    }

    const unsubscribe = ws.subscribe(handleEvent)
    ws.resetReconnectAttempts()

    // 自动连接
    ;(async () => {
      const ready = await ensureGatewayRunning()
      if (!ready) {
        return
      }
      ws.connect().catch((err) => {
        setError(toErrorMessage(err, '连接失败'))
        console.error('WebSocket connection failed:', err)
      })
    })()

    return () => {
      unsubscribe()
      clearReplyTimeout()
    }
  }, [clearReplyTimeout, ensureGatewayRunning])

  // 发送消息
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
    startReplyTimeout()

    wsRef.current.send(content.trim())
  }, [startReplyTimeout])

  // 清空消息
  const clearMessages = useCallback(() => {
    setMessages([])
  }, [])

  // 手动重连
  const reconnect = useCallback(() => {
    setError(null)
    setIsTyping(false)
    clearReplyTimeout()
    wsRef.current.disconnect()
    wsRef.current.resetReconnectAttempts()

    ;(async () => {
      const ready = await ensureGatewayRunning()
      if (!ready) {
        return
      }
      wsRef.current.connect().catch((err) => {
        setError(toErrorMessage(err, '重新连接失败'))
        console.error('Reconnection failed:', err)
      })
    })()
  }, [clearReplyTimeout, ensureGatewayRunning])

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
