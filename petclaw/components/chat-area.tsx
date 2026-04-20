"use client"

import { useEffect, useRef, useState, type CSSProperties } from "react"
import {
  ChevronDown,
  FileText,
  Lightbulb,
  Mic,
  PenTool,
  Plus,
  RefreshCw,
  Search,
  Send,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { type UseChatResult } from "@/hooks/use-chat"
import { useVoiceInput } from "@/hooks/use-voice-input"
import { cn } from "@/lib/utils"

const suggestions = [
  {
    icon: Lightbulb,
    iconColor: "text-yellow-500",
    title: "React Hooks",
    description: "Explain UseEffect dependency arrays simply.",
  },
  {
    icon: PenTool,
    iconColor: "text-orange-500",
    title: "PR Intro",
    description: "Draft a concise pull request description.",
  },
  {
    icon: Search,
    iconColor: "text-blue-500",
    title: "Flexbox Ref",
    description: "Find the CSS property for centering items.",
  },
  {
    icon: FileText,
    iconColor: "text-orange-400",
    title: "Blog Outline",
    description: "Outline a tech post about AI tools.",
  },
]

interface ChatAreaProps {
  chat: UseChatResult
}

export function ChatArea({ chat }: ChatAreaProps) {
  const [activeTab, setActiveTab] = useState("聊天")
  const [message, setMessage] = useState("")
  const messagesEndRef = useRef<HTMLDivElement>(null)
  // 防止 hydration 不匹配：初始渲染时显示加载中状态
  const [isHydrated, setIsHydrated] = useState(false)
  const {
    isListening,
    isSupported: isVoiceInputSupported,
    toggleListening,
  } = useVoiceInput({
    onResult: (text) => {
      setMessage((prev) => `${prev}${prev ? " " : ""}${text}`.trim())
    },
    onError: (voiceError) => {
      console.warn("[petclaw] voice input error", voiceError)
    },
  })

  const {
    messages,
    isConnected,
    isTyping,
    error,
    sendMessage,
    reconnect,
    clearError,
  } = chat

  useEffect(() => {
    setIsHydrated(true)
  }, [])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  const handleSend = () => {
    if (!message.trim()) {
      return
    }
    sendMessage(message)
    setMessage("")
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleSuggestionClick = (title: string, description: string) => {
    sendMessage(`${title}: ${description}`)
  }

  const hasMessages = isHydrated && messages.length > 0

  return (
    <div
      className="flex-1 flex flex-col bg-background"
      style={{ WebkitAppRegion: "no-drag" } as CSSProperties}
    >
      <header className="flex items-center justify-between px-6 py-4 border-b border-border/70 bg-card">
        <div className="flex items-center gap-2">
          <span
            className={cn(
              "w-2 h-2 rounded-full",
              isConnected ? "bg-green-500" : "bg-red-500",
            )}
          />
          <span className="text-sm text-muted-foreground font-medium">
            {isConnected ? "已连接" : "未连接"}
          </span>
          {!isConnected && (
            <Button
              variant="ghost"
              size="sm"
              onClick={reconnect}
              className="h-6 px-2 text-xs"
            >
              <RefreshCw className="w-3 h-3 mr-1" />
              重连
            </Button>
          )}
        </div>

        <div className="flex items-center bg-muted/80 rounded-full p-1 border border-border/70">
          <button
            onClick={() => setActiveTab("聊天")}
            className={cn(
              "px-4 py-1.5 rounded-full text-sm font-medium transition-colors",
              activeTab === "聊天"
                ? "bg-white text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            聊天
          </button>
          <button
            onClick={() => setActiveTab("工作")}
            className={cn(
              "px-4 py-1.5 rounded-full text-sm font-medium transition-colors",
              activeTab === "工作"
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            工作
          </button>
        </div>

        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">PetClaw AI</span>
          <Button
            variant="outline"
            size="sm"
            className="text-amber-700 border-amber-300 bg-amber-50 hover:bg-amber-100 hover:text-amber-800"
          >
            体验增强版
          </Button>
        </div>
      </header>

      <div className="flex-1 flex flex-col overflow-hidden">
        {hasMessages ? (
          <div className="flex-1 overflow-auto px-6 py-5">
            <div className="max-w-3xl mx-auto space-y-4">
              {error && (
                <div className="flex items-center justify-between gap-3 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                  <span>{error}</span>
                  <button
                    onClick={clearError}
                    className="rounded-md px-2 py-1 text-red-700 hover:bg-red-100"
                  >
                    关闭
                  </button>
                </div>
              )}

              {messages.map((msg) => (
                <div
                  key={msg.id}
                  className={cn(
                    "flex",
                    msg.role === "user" ? "justify-end" : "justify-start",
                  )}
                >
                  <div
                    className={cn(
                      "max-w-[80%] rounded-2xl px-4 py-3",
                      msg.role === "user"
                        ? "bg-amber-900 text-amber-50"
                        : "bg-white border border-border/70 text-foreground shadow-sm",
                    )}
                  >
                    <p className="whitespace-pre-wrap">{msg.content}</p>
                    {msg.streaming && (
                      <span className="inline-block w-2 h-4 ml-1 bg-current animate-pulse" />
                    )}
                  </div>
                </div>
              ))}

              {isTyping && !messages.some((msg) => msg.streaming) && (
                <div className="flex justify-start">
                  <div className="bg-muted rounded-2xl px-4 py-3">
                    <div className="flex items-center gap-1">
                      <span
                        className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce"
                        style={{ animationDelay: "0ms" }}
                      />
                      <span
                        className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce"
                        style={{ animationDelay: "150ms" }}
                      />
                      <span
                        className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce"
                        style={{ animationDelay: "300ms" }}
                      />
                    </div>
                  </div>
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>
          </div>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center px-6">
            <div className="mb-6 relative">
              <div className="w-16 h-16 bg-background rounded-2xl border border-border flex items-center justify-center shadow-sm">
                <svg viewBox="0 0 40 40" className="w-10 h-10" fill="none">
                  <circle cx="12" cy="12" r="4" fill="currentColor" />
                  <circle cx="28" cy="12" r="4" fill="currentColor" />
                  <circle cx="12" cy="28" r="4" fill="currentColor" />
                  <circle cx="28" cy="28" r="4" fill="currentColor" />
                  <circle cx="20" cy="20" r="5" fill="currentColor" />
                </svg>
              </div>
              <div className="absolute -top-1 -right-1">
                <svg
                  viewBox="0 0 24 24"
                  className="w-5 h-5 text-foreground"
                  fill="currentColor"
                >
                  <path d="M12 0L13.5 8.5L22 10L13.5 11.5L12 20L10.5 11.5L2 10L10.5 8.5L12 0Z" />
                </svg>
              </div>
            </div>

            <h1 className="text-3xl font-semibold text-foreground mb-3 tracking-tight">
              今天有什么可以帮你?
            </h1>
            <p className="mb-8 text-sm text-muted-foreground">
              可以让桌宠帮你处理日常任务、写作、代码与计划安排。
            </p>

            {error && (
              <div className="mb-4 px-4 py-2 bg-red-50 text-red-600 rounded-lg text-sm">
                {error}
              </div>
            )}

            <div className="grid grid-cols-2 gap-4 max-w-2xl w-full">
              {suggestions.map((item, index) => {
                const Icon = item.icon
                return (
                  <button
                    key={index}
                    onClick={() =>
                      handleSuggestionClick(item.title, item.description)
                    }
                    className="flex items-start gap-3 p-4 rounded-2xl border border-border/80 bg-white hover:bg-accent/60 transition-colors text-left shadow-sm"
                  >
                    <Icon className={cn("w-5 h-5 mt-0.5", item.iconColor)} />
                    <div>
                      <p className="font-medium text-foreground">{item.title}</p>
                      <p className="text-sm text-muted-foreground mt-0.5">
                        {item.description}
                      </p>
                    </div>
                  </button>
                )
              })}
            </div>
          </div>
        )}
      </div>

      <div className="px-6 pb-6 pt-2">
        <div className="max-w-3xl mx-auto">
          <div className="border border-border/80 rounded-2xl bg-white shadow-[0_10px_30px_-20px_rgba(0,0,0,0.45)]">
            <div className="px-4 py-3">
              <input
                type="text"
                placeholder="输入消息..."
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={handleKeyDown}
                className="w-full bg-transparent outline-none text-foreground placeholder:text-muted-foreground"
              />
            </div>

            <div className="flex items-center justify-between px-3 pb-3">
              <div className="flex items-center gap-2">
                <button className="p-2 rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground">
                  <Plus className="w-4 h-4" />
                </button>
                <button className="flex items-center gap-1 px-2 py-1 rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground text-sm">
                  <span>快速</span>
                  <ChevronDown className="w-3 h-3" />
                </button>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={toggleListening}
                  disabled={!isVoiceInputSupported}
                  className={cn(
                    "p-2 rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40",
                    isListening && "bg-accent text-foreground",
                  )}
                  title={
                    isVoiceInputSupported
                      ? isListening
                        ? "停止语音输入"
                        : "开始语音输入"
                      : "当前环境不支持语音输入"
                  }
                >
                  <Mic className="w-4 h-4" />
                </button>
                <Button
                  size="icon"
                  onClick={handleSend}
                  disabled={!message.trim() || !isConnected}
                  className="rounded-full bg-amber-700 text-amber-50 hover:bg-amber-800 w-9 h-9 disabled:opacity-50"
                >
                  {isTyping ? (
                    <Spinner className="w-4 h-4" />
                  ) : (
                    <Send className="w-4 h-4" />
                  )}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
