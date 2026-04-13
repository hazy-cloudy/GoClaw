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
  if (window.runtime) {
    window.runtime.EventsEmit('frontend-log', '[PicoClaw] ' + fullMsg)
  }
}

export function usePicoClaw(apiBaseUrl: string) {
  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef<Record<string, MessageHandler>>({})
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttempts = useRef(0)
  const maxReconnectAttempts = 5
  const apiBaseRef = useRef(apiBaseUrl)
  const sessionIdRef = useRef<string>('')

  useEffect(() => {
    apiBaseRef.current = apiBaseUrl
  }, [apiBaseUrl])

  const fetchTokenInfo = useCallback(async (): Promise<PicoTokenInfo | null> => {
    // Pet channel 配置 - 使用固定配置
    const defaultWsUrl = 'ws://127.0.0.1:8080/ws'
    
    // 优先使用本地配置的 ws_url
    const localWsUrl = import.meta.env.VITE_PICOCLAW_WS_URL || defaultWsUrl
    
    if (localWsUrl) {
      log('使用 Pet Channel:', { ws_url: localWsUrl })
      return {
        token: '',
        ws_url: localWsUrl,
        enabled: true
      }
    }
    
    return null
  }, [])

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return
    
    setIsConnecting(true)
    
    fetchTokenInfo().then(tokenInfo => {
      if (!tokenInfo) {
        log('无法获取连接信息')
        setIsConnecting(false)
        return
      }

      if (!tokenInfo.enabled) {
        log('Pico Channel 未启用')
        setIsConnecting(false)
        return
      }

      log('正在连接...', tokenInfo.ws_url)
      
      try {
        // Pet channel 不需要子协议
        const ws = new WebSocket(tokenInfo.ws_url)
      
      ws.onopen = () => {
        log('✅ 连接成功')
        setIsConnected(true)
        setIsConnecting(false)
        reconnectAttempts.current = 0
      }

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          log('收到消息:', msg.type || 'no-type', 'raw:', JSON.stringify(msg).substring(0, 200))
          
          // Handle session.info from server
          if (msg.type === 'session.info') {
            log('Session info:', msg.session_id)
            sessionIdRef.current = msg.session_id || ''
            return
          }
          
          // Handle push messages
          if (msg.type === 'push') {
            log('推送类型:', msg.push_type)
            
            // 处理 init_status 推送 (v2.0 新增)
            if (msg.push_type === 'init_status') {
              log('init_status:', msg.data)
              const handler = handlersRef.current['init_status']
              if (handler) {
                handler(msg.data)
              }
              return
            }
            
            // 处理 audio 推送 (TTS 音频)
            if (msg.push_type === 'audio') {
              log('收到音频推送:', msg.data)
              const handler = handlersRef.current['audio']
              if (handler) {
                handler(msg.data)
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
              
              if (typeof rawData === 'string') {
                try {
                  const parsed = JSON.parse(rawData)
                  text = parsed.Text || parsed.text || parsed.content || rawData
                  emotion = parsed.Emotion || parsed.emotion || 'neutral'
                  action = parsed.Action || parsed.action || ''
                } catch {
                  text = rawData
                }
              } else if (rawData && typeof rawData === 'object') {
                const dataType = rawData.ContentType
                
                if (dataType === 'text') {
                  text = rawData.Text || ''
                  emotion = rawData.Emotion || 'neutral'
                  action = rawData.Action || ''
                } else if (dataType === 'final') {
                  text = ''
                  emotion = rawData.Emotion || 'neutral'
                  action = rawData.Action || ''
                  isFinal = true
                } else {
                  text = rawData.Text || rawData.text || rawData.content || ''
                  emotion = rawData.Emotion || rawData.emotion || 'neutral'
                  action = rawData.Action || rawData.action || ''
                }
              }
              
              // 清理 text
              if (typeof text === 'string') {
                text = text.replace(/\{[^}]*\}/g, '').trim()
                if (text.startsWith('"') && text.endsWith('"')) {
                  text = text.slice(1, -1)
                }
              }
              
              log('AI回复(push): text=', text, 'emotion=', emotion, 'action=', action, 'isFinal=', isFinal)
              const handler = handlersRef.current['ai_chat']
              if (handler) {
                handler({ text, emotion, action, isFinal })
              }
              return
            }
            
            return
          }
          
          // Handle message.create (pico channel format)
          if (msg.type === 'message.create') {
            log('message.create:', msg.payload)
            let content = msg.payload?.content || ''
            
            // 清理 content - 移除 [text:xxx] 格式
            const textMatch = content.match(/\[text:([^\]]+)\]/)
            if (textMatch) {
              content = textMatch[1]
            }
            
            if (content) {
              log('AI回复(message.create): text=', content)
              const handler = handlersRef.current['ai_chat']
              if (handler) {
                handler({ text: content, emotion: 'neutral', action: '', isFinal: true })
              }
            }
            return
          }
        } catch (e) {
          log('解析消息错误:', String(e))
        }
      }

      ws.onclose = (event) => {
        log('❌ 连接关闭', event.code, event.reason)
        setIsConnected(false)
        wsRef.current = null
        
        // Auto reconnect
        if (reconnectAttempts.current < maxReconnectAttempts) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 30000)
          log(`${delay/1000}秒后重连... (尝试 ${reconnectAttempts.current + 1}/${maxReconnectAttempts})`)
          reconnectTimeoutRef.current = setTimeout(() => {
            reconnectAttempts.current++
            connect()
          }, delay)
        }
      }

      ws.onerror = (error) => {
        log('❌ 连接错误:', String(error))
      }

      wsRef.current = ws
      } catch (e) {
        log('❌ 连接失败:', String(e))
        setIsConnecting(false)
      }
    })
  }, [fetchTokenInfo])

  const disconnect = useCallback(() => {
    log('主动断开连接')
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
        log('WebSocket 未连接，无法发送')
        reject(new Error('WebSocket not connected'))
        return
      }

      const requestId = `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      
      // emotion_get 需要等待响应
      if (action === 'emotion_get') {
        log('emotion_get: waiting for response')
        let resolved = false
        const handler = (event: MessageEvent) => {
          try {
            const msg = JSON.parse(event.data)
            // 服务器返回的是 action，不是 request_id
            if ((msg.request_id === requestId || msg.action === 'emotion_get') && !resolved) {
              log('emotion_get: matched! data:', JSON.stringify(msg.data).substring(0, 100))
              resolved = true
              wsRef.current?.removeEventListener('message', handler)
              resolve(msg.data || msg)
            }
          } catch (e) {
            // ignore parse error
          }
        }
        wsRef.current.addEventListener('message', handler)
        setTimeout(() => {
          if (!resolved) {
            wsRef.current?.removeEventListener('message', handler)
            log('emotion_get: timeout')
            reject(new Error('emotion_get timeout'))
          }
        }, 10000)
        
        const message = {
          action: 'emotion_get',
          data: {},
          request_id: requestId
        }
        log('发送消息:', message.action, message.request_id)
        wsRef.current.send(JSON.stringify(message))
      } else if (action === 'chat') {
        // Pet Channel 消息格式
        const message = {
          action: 'chat',
          data: {
            text: data.text || data.content || '',
            session_key: data.session_key || 'pet:default:go-claw'
          },
          request_id: requestId
        }
        log('发送消息:', message.action, message.request_id)
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
        log('发送消息:', message.action, message.request_id)
        wsRef.current.send(JSON.stringify(message))
        resolve({ status: 'ok' })
      } else if (action === 'health_check') {
        const message = {
          action: 'health_check',
          request_id: requestId
        }
        log('发送消息:', message.action, message.request_id)
        wsRef.current.send(JSON.stringify(message))
        resolve({ status: 'ok' })
      } else {
        // 其他 action
        const message = {
          action: action,
          data: data,
          request_id: requestId
        }
        log('发送消息:', message.action, message.request_id)
        wsRef.current.send(JSON.stringify(message))
        resolve({ status: 'ok' })
      }
    })
  }, [])

  const on = useCallback((pushType: string, handler: MessageHandler) => {
    log('注册处理器:', pushType)
    handlersRef.current[pushType] = handler
  }, [])

  const off = useCallback((pushType: string) => {
    log('移除处理器:', pushType)
    delete handlersRef.current[pushType]
  }, [])

  // Auto connect on mount
  useEffect(() => {
    log('初始化连接，API:', apiBaseUrl)
    connect()
    
    // 注册 heartbeat 处理（主动关怀功能）
    handlersRef.current['heartbeat'] = (data: any) => {
      log('收到 heartbeat:', data)
      // 可选：可以触发一些桌宠的小动作或提示
    }
    
    // Pet channel 使用 WebSocket 标准 ping/pong，不需要额外健康检查
    // 定期发送 ping 保持连接 (WebSocket 标准)
    const healthCheckInterval = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        log('WebSocket ping')
        wsRef.current.send(JSON.stringify({ action: 'health_check', request_id: `ping_${Date.now()}` }))
      }
    }, 30000) // 每30秒发送一次
    
    return () => {
      log('清理连接')
      clearInterval(healthCheckInterval)
      delete handlersRef.current['heartbeat']
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