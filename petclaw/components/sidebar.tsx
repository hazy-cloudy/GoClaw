"use client"

import { useEffect, useRef, useState, type CSSProperties } from "react"
import {
  AlertCircle,
  Clock,
  Cpu,
  FolderUp,
  Home,
  Layers,
  MessageSquare,
  Moon,
  Plus,
  Settings,
  Sparkles,
  Sun,
  Tag,
  User,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { type UseChatResult } from "@/hooks/use-chat"
import { useGatewayStatus } from "@/hooks/use-picoclaw"
import {
  openOnboardingPopup,
  SCHEDULE_ICS_NAME_STORAGE_KEY,
} from "@/lib/onboarding"
import { cn } from "@/lib/utils"

interface SidebarProps {
  activeNav: string
  setActiveNav: (nav: string) => void
  chat: Pick<
    UseChatResult,
    "sessions" | "activeSessionId" | "newChat" | "switchSession"
  >
}

const navItems = [
  { icon: MessageSquare, label: "聊天" },
  { icon: Layers, label: "技能" },
  { icon: Clock, label: "定时任务" },
  { icon: Settings, label: "配置" },
]

function formatRelativeTime(timestamp: number): string {
  const diffMs = Date.now() - timestamp
  if (diffMs < 60_000) return "刚刚"
  if (diffMs < 3_600_000) return `${Math.max(1, Math.floor(diffMs / 60_000))}分钟前`
  if (diffMs < 86_400_000) return `${Math.max(1, Math.floor(diffMs / 3_600_000))}小时前`
  return `${Math.max(1, Math.floor(diffMs / 86_400_000))}天前`
}

export function Sidebar({ activeNav, setActiveNav, chat }: SidebarProps) {
  const [isDarkMode, setIsDarkMode] = useState(false)
  const [scheduleFileName, setScheduleFileName] = useState("")
  const [scheduleImportHint, setScheduleImportHint] = useState("")
  const { data: gatewayStatus } = useGatewayStatus()
  const scheduleFileInputRef = useRef<HTMLInputElement | null>(null)

  useEffect(() => {
    try {
      const savedName = window.localStorage.getItem(SCHEDULE_ICS_NAME_STORAGE_KEY)
      if (savedName) {
        setScheduleFileName(savedName)
      }
    } catch {}
  }, [])

  const toggleTheme = () => {
    setIsDarkMode((prev) => !prev)
    document.documentElement.classList.toggle("dark")
  }

  const getStatusColor = () => {
    if (!gatewayStatus) return "bg-gray-400"
    switch (gatewayStatus.state) {
      case "running":
        return "bg-green-500"
      case "starting":
      case "restarting":
        return "bg-yellow-500 animate-pulse"
      case "stopping":
        return "bg-orange-500"
      default:
        return "bg-red-500"
    }
  }

  const getStatusText = () => {
    if (!gatewayStatus) return "检查中..."
    switch (gatewayStatus.state) {
      case "running":
        return "运行中"
      case "starting":
        return "启动中"
      case "restarting":
        return "重启中"
      case "stopping":
        return "停止中"
      default:
        return "已停止"
    }
  }

  const handleScheduleFilePick = (file: File | null) => {
    if (!file) return
    if (!file.name.toLowerCase().endsWith(".ics")) {
      setScheduleImportHint("只支持 .ics 课表文件")
      return
    }

    setScheduleImportHint("已导入，可在初始化中继续使用")
    setScheduleFileName(file.name)
    try {
      window.localStorage.setItem(SCHEDULE_ICS_NAME_STORAGE_KEY, file.name)
    } catch {}
  }

  return (
    <aside
      className="w-60 border-r border-border/80 bg-sidebar flex flex-col h-full"
      style={{ WebkitAppRegion: "no-drag" } as CSSProperties}
    >
      <div className="p-4 flex items-center gap-2 border-b border-border/60">
        <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center shadow-sm">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            className="w-4 h-4 text-white"
            stroke="currentColor"
            strokeWidth="2"
          >
            <circle cx="7" cy="7" r="2" fill="white" />
            <circle cx="17" cy="7" r="2" fill="white" />
            <circle cx="7" cy="17" r="2" fill="white" />
            <circle cx="17" cy="17" r="2" fill="white" />
            <circle cx="12" cy="12" r="3" fill="white" />
          </svg>
        </div>
        <span className="font-semibold text-foreground tracking-tight">
          PetClaw AI
        </span>
        <span className="text-xs px-1.5 py-0.5 rounded border border-border text-muted-foreground">
          测试版
        </span>
      </div>

      <div className="px-3 my-3">
        <div className="flex items-center gap-2 px-3 py-2.5 rounded-xl bg-card border border-border/70">
          <Cpu className="w-4 h-4 text-muted-foreground" />
          <div className="flex-1 min-w-0">
            <p className="text-xs text-muted-foreground">网关状态</p>
            <div className="flex items-center gap-1.5">
              <span className={cn("w-1.5 h-1.5 rounded-full", getStatusColor())} />
              <span className="text-sm text-foreground truncate">
                {getStatusText()}
              </span>
            </div>
          </div>
          {gatewayStatus?.restartRequired && (
            <span title="需要重启">
              <AlertCircle className="w-4 h-4 text-orange-500" />
            </span>
          )}
        </div>
      </div>

      <div className="px-3 mb-2">
        <Button
          variant="outline"
          className="w-full justify-center gap-2 h-10 bg-background/90 hover:bg-accent border-border/80"
          onClick={() => {
            setActiveNav("聊天")
            void chat.newChat()
          }}
        >
          <Plus className="w-4 h-4" />
          <span>新建聊天</span>
        </Button>
      </div>

      <div className="px-3 mb-3">
        <Button
          variant="secondary"
          className="w-full justify-center gap-2 h-9 bg-gradient-to-r from-amber-100 to-orange-100 text-amber-900 hover:from-amber-200 hover:to-orange-200"
          onClick={openOnboardingPopup}
        >
          <Sparkles className="w-4 h-4" />
          <span>重新初始化</span>
        </Button>
      </div>

      <div className="px-3 mb-3">
        <Button
          variant="outline"
          className="w-full justify-center gap-2 h-9 bg-background/90 hover:bg-accent border-border/80"
          onClick={() => scheduleFileInputRef.current?.click()}
        >
          <FolderUp className="w-4 h-4" />
          <span>导入课表</span>
        </Button>
        <input
          ref={scheduleFileInputRef}
          type="file"
          accept=".ics"
          className="hidden"
          onChange={(event) =>
            handleScheduleFilePick(event.target.files?.[0] ?? null)
          }
        />
        {scheduleImportHint ? (
          <p className="mt-1 px-1 text-xs text-muted-foreground">
            {scheduleImportHint}
          </p>
        ) : scheduleFileName ? (
          <p className="mt-1 px-1 text-xs text-muted-foreground">
            已导入：{scheduleFileName}
          </p>
        ) : null}
      </div>

      <nav className="flex-1 px-3 overflow-auto py-1">
        <ul className="space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = activeNav === item.label
            return (
              <li key={item.label}>
                <button
                  onClick={() => setActiveNav(item.label)}
                  className={cn(
                    "w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-colors",
                    isActive
                      ? "bg-amber-100 text-amber-900"
                      : "text-muted-foreground hover:bg-accent hover:text-foreground",
                  )}
                >
                  <Icon className="w-4 h-4" />
                  <span>{item.label}</span>
                </button>
              </li>
            )
          })}
        </ul>

        <div className="mt-6">
          <p className="text-xs text-muted-foreground px-3 mb-2">对话记录</p>
          <div className="space-y-1">
            {chat.sessions.length === 0 ? (
              <div className="px-3 py-2 text-xs text-muted-foreground">
                暂无历史对话
              </div>
            ) : (
              chat.sessions.map((session) => (
                <button
                  key={session.id}
                  onClick={() => {
                    setActiveNav("聊天")
                    void chat.switchSession(session.id)
                  }}
                  className={cn(
                    "w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm text-left transition-colors",
                    session.id === chat.activeSessionId
                      ? "bg-amber-100 text-amber-900"
                      : "text-muted-foreground hover:bg-accent hover:text-foreground",
                  )}
                >
                  <Home className="w-4 h-4 shrink-0" />
                  <div className="min-w-0 flex-1 flex flex-col items-start">
                    <span className="w-full truncate text-foreground">
                      {session.title}
                    </span>
                    <span className="w-full truncate text-xs text-muted-foreground">
                      {session.preview || formatRelativeTime(session.updatedAt)}
                    </span>
                  </div>
                </button>
              ))
            )}
          </div>
        </div>
      </nav>

      <div className="p-3 border-t border-border/70 bg-card">
        <button className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-muted-foreground hover:bg-accent hover:text-foreground mb-2">
          <Tag className="w-4 h-4" />
          <span>邀请码</span>
        </button>
        <div className="flex items-center justify-between px-2">
          <button className="p-2 rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground">
            <User className="w-4 h-4" />
          </button>
          <button
            onClick={() => setActiveNav("配置")}
            className="p-2 rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground"
          >
            <Settings className="w-4 h-4" />
          </button>
          <button
            onClick={toggleTheme}
            className="p-2 rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground"
          >
            {isDarkMode ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
          </button>
        </div>
      </div>
    </aside>
  )
}
