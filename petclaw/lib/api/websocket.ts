import {
  API_ENDPOINTS,
  getApiBaseUrl,
  getAuthRequestCredentials,
  getWsBaseUrl,
  withLauncherAuthHeader,
} from './config'

const CHAT_ACTION = 'chat'

const PUSH_TYPE_AI_CHAT = 'ai_chat'
const PUSH_TYPE_EMOTION_CHANGE = 'emotion_change'
const PUSH_TYPE_ACTION_TRIGGER = 'action_trigger'

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
  streaming?: boolean
  error?: string
}

interface PetRequest {
  action: string
  data?: Record<string, unknown>
  request_id?: string
}

interface PetResponse {
  status: string
  action?: string
  data?: Record<string, unknown>
  error?: string
  request_id?: string
}

interface PetPush {
  type: string
  push_type: string
  data?: Record<string, unknown>
  timestamp: number
}

interface TokenResponse {
  enabled?: boolean
  token?: string
  ws_url?: string
}

type WSMode = 'pet'
type OutboundRequest = PetRequest

export type WSEventType =
  | 'connected'
  | 'disconnected'
  | 'message'
  | 'typing'
  | 'error'
  | 'reconnecting'
  | 'emotion_change'
  | 'action_trigger'

export interface WSEvent {
  type: WSEventType
  data?: ChatMessage | string
}

type WSEventHandler = (event: WSEvent) => void

export class PicoClawWebSocket {
  private ws: WebSocket | null = null
  private sessionId = ''
  private reconnectAttempts = 0
  private readonly maxReconnectAttempts = 5
  private readonly reconnectDelay = 1000
  private handlers: Set<WSEventHandler> = new Set()
  private wsMode: WSMode = 'pet'
  private messageQueue: OutboundRequest[] = []
  private isConnecting = false
  private msgIdCounter = 0
  private lastConnectUrl = ''
  private manualClose = false
  private openHandlers: {
    settle: () => void
    fail: (err: Error) => void
  } | null = null

  async connect(): Promise<void> {
    return new Promise(async (resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      if (this.isConnecting) {
        const startedAt = Date.now()
        const timer = window.setInterval(() => {
          if (this.ws?.readyState === WebSocket.OPEN) {
            window.clearInterval(timer)
            resolve()
            return
          }
          if (!this.isConnecting || Date.now() - startedAt > 10000) {
            window.clearInterval(timer)
            reject(new Error('Connection in progress timed out'))
          }
        }, 100)
        return
      }

      this.openHandlers = {
        settle: () => resolve(),
        fail: (err: Error) => reject(err),
      }

      try {
        if (!this.sessionId) {
          this.sessionId = this.generateSessionId()
        }
        const { token, wsPath, mode } = await this.resolveTokenAndPath()
        this.wsMode = mode
        const query = `session=${encodeURIComponent(this.sessionId)}&session_id=${encodeURIComponent(this.sessionId)}`
        const url = `${getWsBaseUrl()}${wsPath}?${query}`
        this.connectWebSocket(url, token, mode)
      } catch (err) {
        this.openHandlers = null
        reject(err instanceof Error ? err : new Error('WebSocket bootstrap failed'))
      }
    })
  }

  private async resolveTokenAndPath(): Promise<{ token: string; wsPath: string; mode: WSMode }> {
    const endpoint = { tokenPath: API_ENDPOINTS.PET.TOKEN, wsPath: API_ENDPOINTS.CHAT.WS }

    const res = await fetch(`${getApiBaseUrl()}${endpoint.tokenPath}`, {
      method: 'GET',
      headers: withLauncherAuthHeader(),
      credentials: getAuthRequestCredentials(`${getApiBaseUrl()}${endpoint.tokenPath}`),
    })

    if (res.status === 404) {
      throw new Error('PET channel not available')
    }

    if (!res.ok) {
      throw new Error(`Token endpoint failed (${endpoint.tokenPath}): HTTP ${res.status}`)
    }

    const data = (await res.json()) as TokenResponse
    if (!data.enabled) {
      throw new Error(`Channel not enabled (${endpoint.tokenPath})`)
    }
    const wsPathFromToken = this.normalizeWsPath(data.ws_url)
    const wsPath = wsPathFromToken || endpoint.wsPath
    const mode: WSMode = 'pet'
    const token = data.token || ''

    return {
      token,
      wsPath,
      mode,
    }
  }

  private normalizeWsPath(raw?: string): string {
    if (!raw || !raw.trim()) {
      return ''
    }

    const value = raw.trim()

    try {
      if (value.startsWith('ws://') || value.startsWith('wss://') || value.startsWith('http://') || value.startsWith('https://')) {
        const parsed = new URL(value)
        return parsed.pathname || ''
      }
    } catch {
      return ''
    }

    return value.startsWith('/') ? value : `/${value}`
  }

  private connectWebSocket(url: string, token: string, mode: WSMode): void {
    this.isConnecting = true
    this.lastConnectUrl = url
    this.manualClose = false

    try {
      const protocols = token ? [`token.${token}`] : undefined
      this.ws = protocols ? new WebSocket(url, protocols) : new WebSocket(url)

      this.ws.onopen = () => {
        this.isConnecting = false
        this.reconnectAttempts = 0
        this.emit({ type: 'connected' })

        while (this.messageQueue.length > 0) {
          const msg = this.messageQueue.shift()
          if (msg) this.sendRaw(msg)
        }

        this.openHandlers?.settle()
        this.openHandlers = null
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)

          if (data.type === 'push') {
            this.handlePush(data)
          } else {
            this.handleResponse(data)
          }
        } catch {
          this.emit({ type: 'error', data: 'Invalid message format' })
        }
      }

      this.ws.onerror = () => {
        this.isConnecting = false
        const msg = `WebSocket error (${this.lastConnectUrl})`
        this.emit({ type: 'error', data: msg })
        this.openHandlers?.fail(new Error(msg))
        this.openHandlers = null
      }

      this.ws.onclose = (event) => {
        this.isConnecting = false
        this.emit({ type: 'disconnected' })

        if (this.manualClose) {
          this.manualClose = false
          return
        }

        if (event.code !== 1000) {
          this.emit({ type: 'error', data: `WebSocket closed: code=${event.code} reason=${event.reason || 'none'}` })
        }

        this.attemptReconnect()
      }
    } catch (error) {
      this.isConnecting = false
      const err = error instanceof Error ? error : new Error('WebSocket connection failed')
      this.openHandlers?.fail(err)
      this.openHandlers = null
    }
  }

  private handlePush(push: PetPush): void {
    const data = push.data || {}

    switch (push.push_type) {
      case PUSH_TYPE_AI_CHAT:
        this.handleAIChatPush(data)
        break
      case PUSH_TYPE_EMOTION_CHANGE:
        this.handleEmotionChangePush(data)
        break
      case PUSH_TYPE_ACTION_TRIGGER:
        this.handleActionTriggerPush(data)
        break
    }
  }

  private handleAIChatPush(data: Record<string, unknown>): void {
    const chatId = (data.chat_id as number) || 0
    const contentType = (data.type as string) || 'text'
    const text = (data.text as string) || ''

    if (contentType === 'text' || contentType === '') {
      const chatMsg: ChatMessage = {
        id: `msg-${chatId}`,
        role: 'assistant',
        content: text,
        timestamp: Date.now(),
        streaming: true,
      }
      this.emit({ type: 'message', data: chatMsg })
      return
    }

    if (contentType === 'final') {
      const chatMsg: ChatMessage = {
        id: `msg-${chatId}`,
        role: 'assistant',
        content: text,
        timestamp: Date.now(),
        streaming: false,
      }
      this.emit({ type: 'message', data: chatMsg })
      this.emit({ type: 'typing', data: 'false' })
    }
  }

  private handleEmotionChangePush(data: Record<string, unknown>): void {
    const emotion = (data.emotion as string) || ''
    this.emit({ type: 'emotion_change', data: emotion })
  }

  private handleActionTriggerPush(data: Record<string, unknown>): void {
    const action = (data.action as string) || ''
    this.emit({ type: 'action_trigger', data: action })
  }

  private handleResponse(_resp: PetResponse): void {
  }

  disconnect(): void {
    if (this.ws) {
      this.manualClose = true
      this.ws.close()
      this.ws = null
    }
    this.reconnectAttempts = 0
  }

  resetReconnectAttempts(): void {
    this.reconnectAttempts = 0
  }

  send(content: string): void {
    this.sendAction(CHAT_ACTION, {
      text: content,
      session_key: this.sessionId,
    })
  }

  private sendAction(action: string, data?: Record<string, unknown>): void {
    const requestId = `req-${++this.msgIdCounter}-${Date.now()}`
    const msg: PetRequest = {
      action,
      data,
      request_id: requestId,
    }

    if (this.ws?.readyState === WebSocket.OPEN) {
      this.sendRaw(msg)
      return
    }

    this.messageQueue.push(msg)
    this.connect().catch(() => {
      this.emit({ type: 'error', data: 'Connection failed' })
    })
  }

  private sendRaw(msg: OutboundRequest): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
    }
  }

  subscribe(handler: WSEventHandler): () => void {
    this.handlers.add(handler)
    return () => this.handlers.delete(handler)
  }

  private emit(event: WSEvent): void {
    this.handlers.forEach((handler) => handler(event))
  }

  private generateSessionId(): string {
    return `session-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.emit({ type: 'error', data: `Max reconnection attempts reached (${this.lastConnectUrl || 'unknown url'})` })
      return
    }

    this.reconnectAttempts++
    this.emit({ type: 'reconnecting' })

    window.setTimeout(() => {
      this.connect().catch(() => {})
    }, this.reconnectDelay * this.reconnectAttempts)
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

let wsInstance: PicoClawWebSocket | null = null

export function getWebSocketInstance(): PicoClawWebSocket {
  if (!wsInstance) {
    wsInstance = new PicoClawWebSocket()
  }
  return wsInstance
}
