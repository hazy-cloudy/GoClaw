"use client"

import { Minus, Square, X } from "lucide-react"
import { useEffect, useState, type CSSProperties } from "react"

interface WindowControlsProps {
  title: string
}

export function WindowControls({ title }: WindowControlsProps) {
  const [hasElectronApi, setHasElectronApi] = useState(false)

  useEffect(() => {
    setHasElectronApi(Boolean(window.electronAPI))
  }, [])

  return (
    <div className="z-20 flex h-10 select-none items-center justify-between border-b border-border/70 bg-card px-3" style={{ userSelect: "none" } as CSSProperties}>
      <div className="min-w-0 flex-1 truncate text-xs font-medium text-muted-foreground" data-electron-drag-region="true">
        {title}
      </div>
      <div className="ml-2 flex items-center gap-1" data-electron-no-drag="true">
        <button
          type="button"
          onClick={() => window.electronAPI?.minimizeWindow?.()}
          disabled={!hasElectronApi}
          className="rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
          title="最小化"
        >
          <Minus className="h-3.5 w-3.5" />
        </button>
        <button
          type="button"
          onClick={() => window.electronAPI?.toggleMaximizeWindow?.()}
          disabled={!hasElectronApi}
          className="rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
          title="最大化 / 还原"
        >
          <Square className="h-3.5 w-3.5" />
        </button>
        <button
          type="button"
          onClick={() => {
            if (hasElectronApi) {
              window.electronAPI?.closeWindow?.()
              return
            }
            if (typeof window !== "undefined") {
              window.close()
            }
          }}
          className="rounded-md p-1.5 text-muted-foreground hover:bg-red-500 hover:text-white"
          title="关闭"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  )
}
