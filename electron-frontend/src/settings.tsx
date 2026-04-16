import { useState, useEffect, useCallback, useRef } from 'react'
import { usePicoClaw } from './usePicoClaw'
import './settings.css'

interface ChatMessage {
  id: string
  text: string
  time: string
  isUser?: boolean
}

interface Settings {
  character: string
  opacity: number
  speechEnabled: boolean
  voiceVolume: number
  windowScale: number
}

const CHARACTERS = [
  { id: 'lively', name: '活泼', desc: '元气满满' },
  { id: 'healing', name: '治愈', desc: '温暖人心' },
  { id: 'silly', name: '沙雕', desc: '搞笑担当' },
  { id: 'clingy', name: '粘人', desc: '撒娇达人' },
  { id: 'mischievous', name: '腹黑', desc: '腹黑萌' },
  { id: 'foodie', name: '吃货', desc: '美食至上' }
]

type TabType = 'chat' | 'character' | 'llm' | 'display'

function Settings() {
  const [activeTab, setActiveTab] = useState<TabType>('chat')
  const [chatHistory, setChatHistory] = useState<ChatMessage[]>([])
  const chatRef = useRef('')
  const emotionRef = useRef('neutral')
  const audioRef = useRef('')
  const [inputText, setInputText] = useState('')
  const [streamingText, setStreamingText] = useState('')
  const currentAIMsgId = useRef<string | null>(null)
  const [connStatus, setConnStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting')
  const [settings, setSettings] = useState<Settings>({
    character: 'lively',
    opacity: 1,
    speechEnabled: true,
    voiceVolume: 80,
    windowScale: 1
  })

  const PICOCLAW_API_URL = 'http://127.0.0.1:18790'
  const { send: picoSend, on: picoOn, off: picoOff, isConnected, isConnecting } = usePicoClaw(PICOCLAW_API_URL, {
    onConnectionChange: (connected) => {
      if (connected) {
        window.electronAPI?.sendConnectionAlive?.()
      }
    }
  })

  useEffect(() => {
    if (isConnected) {
      setConnStatus('connected')
    } else if (isConnecting) {
      setConnStatus('connecting')
    } else {
      setConnStatus('disconnected')
    }
  }, [isConnected, isConnecting])

  useEffect(() => {
    picoOn('ai_chat', (data) => {
      if (data.is_final) {
        const finalText = chatRef.current
        chatRef.current = ''
        if (finalText) {
          setChatHistory(prev => [...prev, {
            id: `ai_${Date.now()}`,
            text: finalText,
            time: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
            isUser: false
          }])
        }
        return
      }
      if (data.text) {
        if (data.emotion) emotionRef.current = data.emotion
        chatRef.current += data.text
      }
    })
    
    picoOn('audio', (data) => {
      if (data.text) {
        audioRef.current += data.text
      }
      if (data.is_final && audioRef.current) {
        window.electronAPI?.showBubble(null, emotionRef.current, audioRef.current)
        audioRef.current = ''
      }
    })
    
    return () => {
      picoOff('ai_chat')
      picoOff('audio')
    }
  }, [picoOn, picoOff])

  useEffect(() => {
    const saved = localStorage.getItem('go-claw-settings')
    if (saved) {
      setSettings(JSON.parse(saved))
    }
  }, [])

  const handleClose = () => {
    window.electronAPI?.closeSettings()
  }

  const handleMinimize = () => {
    window.electronAPI?.minimizeSettings()
  }

  const handleMaximize = () => {
    window.electronAPI?.maximizeSettings()
  }

  const sendMessage = useCallback(async () => {
    if (!inputText.trim()) return
    const userMsg: ChatMessage = {
      id: `user_${Date.now()}`,
      text: inputText.trim(),
      time: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
      isUser: true
    }
    setChatHistory(prev => [...prev, userMsg])
    setInputText('')
    
    try {
      await picoSend('chat', { text: inputText.trim() })
    } catch (e) {
      console.error('[Settings] send error:', e)
    }
  }, [inputText, picoSend])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      sendMessage()
    }
  }

  const selectCharacter = (id: string) => {
    const newSettings = { ...settings, character: id }
    setSettings(newSettings)
    saveSettings(newSettings)
  }

  const saveSettings = (s: Settings) => {
    localStorage.setItem('go-claw-settings', JSON.stringify(s))
  }

  const updateOpacity = (val: number) => {
    const newSettings = { ...settings, opacity: val / 100 }
    setSettings(newSettings)
    saveSettings(newSettings)
  }

  const updateScale = (val: number) => {
    const newSettings = { ...settings, windowScale: val / 100 }
    setSettings(newSettings)
    saveSettings(newSettings)
  }

  const updateSpeechEnabled = (enabled: boolean) => {
    const newSettings = { ...settings, speechEnabled: enabled }
    setSettings(newSettings)
    saveSettings(newSettings)
  }

  const updateVolume = (vol: number) => {
    const newSettings = { ...settings, voiceVolume: vol }
    setSettings(newSettings)
    saveSettings(newSettings)
  }

  return (
    <div className="settings-app">
      <div className="settings-header">
        <div className="settings-title">
          <span className="settings-icon">⚙️</span>
          <span className="settings-title-text">桌宠设置</span>
        </div>
        <div className={`connection-status ${connStatus}`}>
          <span className="status-dot"></span>
          <span className="status-text">
            {connStatus === 'connected' ? '已连接' : connStatus === 'connecting' ? '连接中' : '未连接'}
          </span>
        </div>
        <div className="settings-controls">
          <button className="settings-btn minimize" onClick={handleMinimize}>─</button>
          <button className="settings-btn maximize" onClick={handleMaximize}>☐</button>
          <button className="settings-btn close" onClick={handleClose}>×</button>
        </div>
      </div>

      <div className="settings-body">
        <div className="settings-tabs">
        <button
          className={`settings-tab ${activeTab === 'chat' ? 'active' : ''}`}
          onClick={() => setActiveTab('chat')}
        >
          <span className="tab-icon">💬</span>
          <span className="tab-name">聊天记录</span>
        </button>
        <button
          className={`settings-tab ${activeTab === 'character' ? 'active' : ''}`}
          onClick={() => setActiveTab('character')}
        >
          <span className="tab-icon">🎭</span>
          <span className="tab-name">人格切换</span>
        </button>
        <button
          className={`settings-tab ${activeTab === 'llm' ? 'active' : ''}`}
          onClick={() => setActiveTab('llm')}
        >
          <span className="tab-icon">🤖</span>
          <span className="tab-name">大模型</span>
        </button>
        <button
          className={`settings-tab ${activeTab === 'display' ? 'active' : ''}`}
          onClick={() => setActiveTab('display')}
        >
          <span className="tab-icon">⚙️</span>
          <span className="tab-name">显示设置</span>
        </button>
      </div>

      <div className="settings-content">
        {activeTab === 'chat' && (
          <div className="settings-page">
            <h3>💬 与桌宠聊天</h3>
            <div className="chat-list">
              {chatHistory.length === 0 ? (
                <div className="chat-empty">开始和桌宠聊天吧...</div>
              ) : (
                chatHistory.map((msg) => (
                  <div key={msg.id} className={`chat-message ${msg.isUser ? 'user' : 'ai'}`}>
                    <div className="message-bubble">
                      <span className="message-text">{msg.text}</span>
                    </div>
                    <span className="message-time">{msg.time}</span>
                  </div>
                ))
              )}
            </div>
            <div className="chat-input-container">
              <input
                type="text"
                className="chat-input"
                placeholder="输入消息..."
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onKeyDown={handleKeyDown}
              />
              <button className="chat-send-btn" onClick={sendMessage}>发送</button>
            </div>
          </div>
        )}

        {activeTab === 'character' && (
          <div className="settings-page">
            <h3>🎭 选择桌宠人格</h3>
            <div className="char-grid">
              {CHARACTERS.map((char) => (
                <button
                  key={char.id}
                  className={`char-btn ${settings.character === char.id ? 'active' : ''}`}
                  onClick={() => selectCharacter(char.id)}
                >
                  <span className="char-name">{char.name}</span>
                  <span className="char-desc">{char.desc}</span>
                </button>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'llm' && (
          <div className="settings-page">
            <h3>🤖 大模型配置</h3>
            <div className="llm-form">
              <div className="form-group">
                <label>API 地址</label>
                <input type="text" defaultValue="http://127.0.0.1:18790" />
              </div>
              <div className="form-group">
                <label>API Key</label>
                <input type="password" placeholder="输入 API Key" />
              </div>
              <div className="form-group">
                <label>模型</label>
                <select defaultValue="minimax/MiniMax-Text-01">
                  <option value="minimax/MiniMax-Text-01">MiniMax-Text-01</option>
                  <option value="minimax/MiniMax-M2.5">MiniMax-M2.5</option>
                  <option value="zhipu/glm-4">GLM-4</option>
                </select>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'display' && (
          <div className="settings-page">
            <h3>⚙️ 显示设置</h3>
            <div className="display-settings">
              <div className="setting-row">
                <label>透明度</label>
                <input
                  type="range"
                  min="30"
                  max="100"
                  value={Math.round(settings.opacity * 100)}
                  onChange={(e) => updateOpacity(parseInt(e.target.value))}
                />
                <span>{Math.round(settings.opacity * 100)}%</span>
              </div>
              <div className="setting-row">
                <label>大小</label>
                <input
                  type="range"
                  min="70"
                  max="120"
                  step="5"
                  value={Math.round(settings.windowScale * 100)}
                  onChange={(e) => updateScale(parseInt(e.target.value))}
                />
                <span>{Math.round(settings.windowScale * 100)}%</span>
              </div>
              <div className="setting-row">
                <label>语音播报</label>
                <label className="switch">
                  <input
                    type="checkbox"
                    checked={settings.speechEnabled}
                    onChange={(e) => updateSpeechEnabled(e.target.checked)}
                  />
                  <span className="slider"></span>
                </label>
              </div>
              <div className="setting-row">
                <label>音量</label>
                <input
                  type="range"
                  min="0"
                  max="100"
                  step="10"
                  value={settings.voiceVolume}
                  onChange={(e) => updateVolume(parseInt(e.target.value))}
                />
                <span>{settings.voiceVolume}%</span>
              </div>
            </div>
          </div>
        )}
        </div>
      </div>
    </div>
  )
}

export default Settings