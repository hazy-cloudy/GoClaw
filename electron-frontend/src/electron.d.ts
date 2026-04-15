export {}

declare global {
  interface Window {
    electronAPI?: {
      openSettings: () => void
      minimizeSettings: () => void
      maximizeSettings: () => void
      closeSettings: () => void
      sendSettingsChange: (settings: any) => void
      sendChatHistory: (history: any[]) => void
      showBubble: (text: string | null, emotion: string, audio?: string) => void
      onSettingsUpdate: (callback: (settings: any) => void) => void
      onChatHistoryUpdate: (callback: (history: any[]) => void) => void
      onBubbleShow: (callback: (data: { text: string | null; emotion?: string; audio?: string }) => void) => void
    }
  }
}