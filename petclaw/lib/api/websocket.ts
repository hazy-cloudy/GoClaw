import {
  API_ENDPOINTS,
  USE_CREDENTIALS,
  getApiBaseUrl,
  getWsBaseUrl,
  withLauncherAuthHeader,
} from './config'
import { fetchWithAuthRetry } from './auth-bootstrap'

interface PicoTokenInfo {
  token: string
  ws_url: string
  enabled: boolean
}

function normalizeWsUrlForBrowser(rawUrl: string): string {
  const base = typeof window !== 'undefined' ? window.location.origin : 'http://127.0.0.1:18800'
  const parsed = new URL(rawUrl, base)

  if (parsed.protocol === 'http:') {
    parsed.protocol = 'ws:'
  } else if (parsed.protocol === 'https:') {
    parsed.protocol = 'wss:'
  }

  return parsed.toString().replace(/\/$/, '')
}

function getFallbackWebSocketURL(): string {
  return `${getWsBaseUrl()}${API_ENDPOINTS.CHAT.WS}`
}

const CHANNEL_BOOTSTRAP_ENDPOINTS = [
  { token: API_ENDPOINTS.PET.TOKEN, setup: API_ENDPOINTS.PET.SETUP, wsPath: API_ENDPOINTS.CHAT.WS },
  { token: API_ENDPOINTS.PICO.TOKEN, setup: API_ENDPOINTS.PICO.SETUP, wsPath: API_ENDPOINTS.CHAT.WS_LEGACY },
]

function forceWsPath(rawUrl: string, wsPath: string): string {
  const parsed = new URL(rawUrl, typeof window !== 'undefined' ? window.location.origin : 'http://127.0.0.1:18800')
  parsed.pathname = wsPath
  return parsed.toString()
}

async function fetchPicoToken(): Promise<PicoTokenInfo> {
  let lastStatus = 0

  for (const endpoint of CHANNEL_BOOTSTRAP_ENDPOINTS) {
    const tokenRes = await fetchWithAuthRetry(`${getApiBaseUrl()}${endpoint.token}`, {
      headers: withLauncherAuthHeader(),
      credentials: USE_CREDENTIALS ? 'include' : 'omit',
    })

    lastStatus = tokenRes.status
    if (!tokenRes.ok) {
      if (tokenRes.status === 404) {
        continue
      }
      throw new Error(`Failed to get channel token: ${tokenRes.status}`)
    }

    const tokenData = (await tokenRes.json()) as PicoTokenInfo
    tokenData.ws_url = tokenData.ws_url
      ? forceWsPath(tokenData.ws_url, endpoint.wsPath)
      : `${getWsBaseUrl()}${endpoint.wsPath}`
    if (tokenData.enabled) {
      return tokenData
    }

    const setupRes = await fetchWithAuthRetry(`${getApiBaseUrl()}${endpoint.setup}`, {
      method: 'POST',
      headers: withLauncherAuthHeader({
        'Content-Type': 'application/json',
      }),
      credentials: USE_CREDENTIALS ? 'include' : 'omit',
    })

    if (!setupRes.ok) {
      if (setupRes.status === 404) {
        continue
      }
      throw new Error(`Failed to setup channel: ${setupRes.status}`)
    }

    const setupData = (await setupRes.json()) as PicoTokenInfo
    setupData.ws_url = setupData.ws_url
      ? forceWsPath(setupData.ws_url, endpoint.wsPath)
      : `${getWsBaseUrl()}${endpoint.wsPath}`
    return setupData
  }

  throw new Error(`Failed to bootstrap channel token: ${lastStatus || 404}`)
}

// Pico Protocol 消息类型
const PICO_TYPE_MESSAGE_SEND = "message.send"
const PICO_TYPE_MESSAGE_CREATE = "message.create"
const PICO_TYPE_MESSAGE_UPDATE = "message.update"
const PICO_TYPE_TYPING_START = "typing.start"
const PICO_TYPE_TYPING_STOP = "typing.stop"
const PICO_TYPE_PING = "ping"
const PICO_TYPE_PONG = "pong"
const PICO_TYPE_ERROR = "error"

// 聊天消息类型
export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
  streaming?: boolean
  error?: string
}

// WebSocket 事件类型
export type WSEventType = 
  | 'connected'
  | 'disconnected'
  | 'message'
  | 'typing'
  | 'error'
  | 'reconnecting'

export interface WSEvent {
  type: WSEventType
  data?: ChatMessage | string
}

type WSEventHandler = (event: WSEvent) => void

export class PicoClawWebSocket {
  private ws: WebSocket | null = null
  private token: string = ''
  private sessionId: string = ''
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private handlers: Set<WSEventHandler> = new Set()
  private messageQueue: any[] = []
  private isConnecting = false
  private msgIdCounter = 0
  private lastConnectUrl = ''
  private manualClose = false

  constructor() {}

  async connect(): Promise<void> {
    return new Promise(async (resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      if (this.isConnecting) {
        const startedAt = Date.now()
        const checkConnection = setInterval(() => {
          if (this.ws?.readyState === WebSocket.OPEN) {
            clearInterval(checkConnection)
            resolve()
            return
          }

          if (!this.isConnecting || Date.now() - startedAt > 10000) {
            clearInterval(checkConnection)
            reject(new Error('Connection in progress timed out'))
          }
        }, 100)
        return
      }

      const settle = () => resolve()
      const fail = (err: Error) => reject(err)
      this.openHandlers = { settle, fail }

      try {
        const pico = await fetchPicoToken()
        this.token = pico.token
        this.sessionId = this.generateSessionId()

        const resolvedWsUrl = pico.ws_url
          ? normalizeWsUrlForBrowser(pico.ws_url)
          : getFallbackWebSocketURL()

        const url = `${resolvedWsUrl}?session_id=${encodeURIComponent(this.sessionId)}`
        this.connectWebSocket(url)
      } catch (err) {
        this.openHandlers = null
        reject(err)
        return
      }
    })
  }

  private openHandlers: {
    settle: () => void
    fail: (err: Error) => void
  } | null = null

  private connectWebSocket(url: string): void {
    this.isConnecting = true
    this.lastConnectUrl = url
    this.manualClose = false

    try {
      this.ws = new WebSocket(url, [`token.${this.token}`])

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
          const picoMsg = JSON.parse(event.data)
          this.handlePicoMessage(picoMsg)
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

  private handlePicoMessage(msg: any): void {
    const payload = (msg && typeof msg.payload === 'object' && msg.payload) || {}

    switch (msg.type) {
      case PICO_TYPE_MESSAGE_CREATE:
        {
          const content = payload['content'] as string
          const isThought = payload['thought'] as boolean
          const messageId =
            (payload['message_id'] as string) ||
            (typeof msg.id === 'string' ? msg.id : '') ||
            `msg-${Date.now()}`

          const chatMsg: ChatMessage = {
            id: messageId,
            role: isThought ? 'system' : 'assistant',
            content: typeof content === 'string' ? content : '',
            timestamp: this.normalizeTimestamp(msg.timestamp),
            streaming: false,
          }
          this.emit({ type: 'message', data: chatMsg })
        }
        break

      case PICO_TYPE_MESSAGE_UPDATE:
        {
          const content = payload['content'] as string
          const messageId =
            (payload['message_id'] as string) ||
            (typeof msg.id === 'string' ? msg.id : '')

          if (!messageId) {
            break
          }

          const chatMsg: ChatMessage = {
            id: messageId,
            role: 'assistant',
            content: typeof content === 'string' ? content : '',
            timestamp: this.normalizeTimestamp(msg.timestamp),
            streaming: true,
          }
          this.emit({ type: 'message', data: chatMsg })
        }
        break

      case PICO_TYPE_TYPING_START:
        this.emit({ type: 'typing', data: 'true' })
        break

      case PICO_TYPE_TYPING_STOP:
        this.emit({ type: 'typing', data: 'false' })
        break

      case PICO_TYPE_PONG:
        break

      case PICO_TYPE_PING:
        this.sendRaw({
          type: PICO_TYPE_PONG,
          id: msg.id,
          session_id: this.sessionId,
          timestamp: Date.now(),
        })
        break

      case PICO_TYPE_ERROR:
        {
          const code = payload['code'] as string
          const errorMsg = payload['message'] as string
          this.emit({ type: 'error', data: `${code}: ${errorMsg}` })
        }
        break
    }
  }

  private normalizeTimestamp(value: unknown): number {
    const num = Number(value)
    if (!Number.isFinite(num)) {
      return Date.now()
    }
    return num < 1_000_000_000_000 ? num * 1000 : num
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
    const id = `msg-${++this.msgIdCounter}-${Date.now()}`
    const msg: any = {
      type: PICO_TYPE_MESSAGE_SEND,
      id,
      session_id: this.sessionId,
      timestamp: Date.now(),
      payload: {
        content: content,
        media: [],
      },
    }

    if (this.ws?.readyState === WebSocket.OPEN) {
      this.sendRaw(msg)
    } else {
      this.messageQueue.push(msg)
      this.connect().catch(() => {
        this.emit({ type: 'error', data: 'Connection failed' })
      })
    }
  }

  private sendRaw(msg: any): void {
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

    setTimeout(() => {
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
