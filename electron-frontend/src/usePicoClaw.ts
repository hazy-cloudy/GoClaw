import { useState, useEffect, useRef, useCallback } from 'react'

interface PicoTokenInfo {
  token: string
  ws_url: string
  enabled: boolean
}

type MessageHandler = (data: any) => void

const log = (msg: string, ...args: any[]) => {
  const fullMsg = args.length > 0 ? `${msg} ${args.map(a => JSON.stringify(a)).join(' ')}` : msg
  console.log('[PicoClaw]', fullMsg)
}

interface PicoCallbacks {
  onHeartbeat?: () => void
  onEmotionChange?: (emotion: string) => void
  onConnectionChange?: (connected: boolean) => void
}

export function usePicoClaw(apiBaseUrl: string, callbacks?: PicoCallbacks) {
  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef<Record<string, MessageHandler>>({})
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttempts = useRef(0)
  const maxReconnectAttempts = 5
  const apiBaseRef = useRef(apiBaseUrl)
  const sessionIdRef = useRef<string>('')
  const callbacksRef = useRef(callbacks)
  callbacksRef.current = callbacks

  useEffect(() => {
    apiBaseRef.current = apiBaseUrl
  }, [apiBaseUrl])

  const fetchTokenInfo = useCallback(async (): Promise<PicoTokenInfo | null> => {
    const wsUrl = 'ws://127.0.0.1:8080/ws'
    return {
      token: '',
      ws_url: wsUrl,
      enabled: true
    }
  }, [])

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return
    
    // 关闭旧连接
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    
    setIsConnecting(true)
    
    fetchTokenInfo().then(tokenInfo => {
      if (!tokenInfo) {
        setIsConnecting(false)
        return
      }

      if (!tokenInfo.enabled) {
        setIsConnecting(false)
        return
      }

      try {
        // Pet channel 不需要子协议
        const ws = new WebSocket(tokenInfo.ws_url)
      
      ws.onopen = () => {
        setIsConnected(true)
        setIsConnecting(false)
        reconnectAttempts.current = 0
        callbacksRef.current?.onConnectionChange?.(true)
      }

      ws.onerror = () => {
        setIsConnected(false)
        setIsConnecting(false)
        callbacksRef.current?.onConnectionChange?.(false)
      }

      ws.onclose = () => {
        setIsConnected(false)
        callbacksRef.current?.onConnectionChange?.(false)
      }

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          
          // Handle session.info from server
          if (msg.type === 'session.info') {
            sessionIdRef.current = msg.session_id || ''
            return
          }
          
          // Handle push messages
          if (msg.type === 'push') {
            
            // 处理 init_status 推送 (v2.0 新增)
            if (msg.push_type === 'init_status') {
              const handler = handlersRef.current['init_status']
              if (handler) {
                handler(msg.data)
              }
              return
            }
            
            // 处理 audio 推送 (TTS 音频)
            if (msg.push_type === 'audio') {
              const handler = handlersRef.current['audio']
              if (handler) {
                let audioData = msg.data
                if (typeof audioData === 'string') {
                  try {
                    audioData = JSON.parse(audioData)
                  } catch (e) {}
                }
                handler(audioData)
              }
              return
            }
            
// 处理 ai_chat 推送 (支持流式 v2.0)
            if (msg.push_type === 'ai_chat') {
              const rawData = msg.data
              let text = ''
              let emotion = 'neutral'
              let action = ''
              let isFinal = msg.is_final === true
              let chatId = 0
              
              if (typeof rawData === 'string') {
                try {
                  const parsed = JSON.parse(rawData)
                  text = parsed.text || parsed.Text || parsed.content || rawData
                  emotion = parsed.emotion || parsed.Emotion || 'neutral'
                  action = parsed.action || parsed.Action || ''
                  chatId = parsed.chat_id || 0
                } catch {
                  text = rawData
                }
              } else if (rawData && typeof rawData === 'object') {
                chatId = rawData.chat_id || rawData.ChatID || rawData.chatId || 0
                const dataType = rawData.type || rawData.ContentType
                
                if (dataType === 'text' || dataType === 'final') {
                  text = rawData.Text || rawData.text || rawData.content || ''
                  emotion = rawData.Emotion || rawData.emotion || 'neutral'
                  action = rawData.Action || rawData.action || ''
                  if (dataType === 'final') isFinal = true
                } else {
                  text = rawData.Text || rawData.text || rawData.content || ''
                  emotion = rawData.Emotion || rawData.emotion || 'neutral'
                  action = rawData.Action || rawData.action || ''
                }
              }
              
              if (typeof text === 'string') {
                text = text.replace(/\{[^}]*\}/g, '').trim()
                if (text.startsWith('"') && text.endsWith('"')) {
                  text = text.slice(1, -1)
                }
              }
               
              const handler = handlersRef.current['ai_chat']
              if (handler) {
                handler({ text, emotion, action, isFinal, chat_id: chatId, is_final: isFinal })
              }
              return
            }
            
            return
          }
          
          // Handle message.create (pico channel format)
          if (msg.type === 'message.create') {
            // 跳过
            return
          }
        } catch (e) {}
      }

      ws.onclose = () => {
        setIsConnected(false)
        wsRef.current = null
        
        // Auto reconnect
        if (reconnectAttempts.current < maxReconnectAttempts) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000)
          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttempts.current++
            connect()
          }, delay)
        }
      }

      ws.onerror = () => {}

      wsRef.current = ws
      } catch (e) {
        setIsConnecting(false)
      }
    })
  }, [fetchTokenInfo])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setIsConnected(false)
  }, [])

  const send = useCallback((action: string, data: any): Promise<any> => {
    return new Promise((resolve, reject) => {
      if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
        reject(new Error('WebSocket not connected'))
        return
      }

      const requestId = `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      
      if (action === 'emotion_get') {
        let resolved = false
        const handler = (event: MessageEvent) => {
          try {
            const msg = JSON.parse(event.data)
            if ((msg.request_id === requestId || msg.action === 'emotion_get') && !resolved) {
              resolved = true
              wsRef.current?.removeEventListener('message', handler)
              resolve(msg.data || msg)
            }
          } catch (e) {}
        }
        wsRef.current.addEventListener('message', handler)
        setTimeout(() => {
          if (!resolved) {
            wsRef.current?.removeEventListener('message', handler)
            reject(new Error('emotion_get timeout'))
          }
        }, 10000)
        
        const message = {
          action: 'emotion_get',
          data: {},
          request_id: requestId
        }
        wsRef.current.send(JSON.stringify(message))
      } else if (action === 'chat') {
        const message = {
          action: 'chat',
          data: {
            text: data.text || data.content || '',
            session_key: data.session_key || 'pet:default:go-claw'
          },
          request_id: requestId
        }
        wsRef.current.send(JSON.stringify(message))
        resolve({ status: 'ok' })
      } else if (action === 'tool_result') {
        // 工具执行结果反馈
        const message = {
          action: 'tool_result',
          data: {
            tool_call_id: data.tool_call_id,
            result: data.result
          },
          request_id: requestId
        }
        wsRef.current.send(JSON.stringify(message))
        resolve({ status: 'ok' })
      } else if (action === 'health_check') {
        const message = {
          action: 'health_check',
          request_id: requestId
        }
        wsRef.current.send(JSON.stringify(message))
        resolve({ status: 'ok' })
      } else {
        const message = {
          action: action,
          data: data,
          request_id: requestId
        }
        wsRef.current.send(JSON.stringify(message))
        resolve({ status: 'ok' })
      }
    })
  }, [])

  const on = useCallback((pushType: string, handler: MessageHandler) => {
    handlersRef.current[pushType] = handler
  }, [])

  const off = useCallback((pushType: string) => {
    delete handlersRef.current[pushType]
  }, [])

  useEffect(() => {
    connect()
    
    handlersRef.current['heartbeat'] = () => {
      if (callbacksRef.current?.onHeartbeat) {
        callbacksRef.current.onHeartbeat()
      }
    }
    
    handlersRef.current['emotion_change'] = (data: any) => {
      const emotion = data?.emotion || data
      if (callbacksRef.current?.onEmotionChange && emotion) {
        callbacksRef.current.onEmotionChange(emotion)
      }
    }
    
    // 注册 audio 处理 (TTS 音频数据)
    handlersRef.current['audio'] = (data: any) => {
      if (data && data.text) {
        window.electronAPI?.showBubble(null, 'happy', data.text)
      }
    }
    
    const healthCheckInterval = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ action: 'health_check', request_id: `ping_${Date.now()}` }))
      }
    }, 30000)
    
    return () => {
      clearInterval(healthCheckInterval)
      delete handlersRef.current['heartbeat']
      delete handlersRef.current['audio']
      disconnect()
    }
  }, [connect, disconnect, apiBaseUrl])

  return {
    isConnected,
    isConnecting,
    connect,
    disconnect,
    send,
    on,
    off
  }
}