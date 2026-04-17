export {}

declare global {
  interface Window {
    electronAPI?: {
      openDashboard: () => void
      openOnboarding: () => void
      setOnboardingMode: (enabled: boolean) => void
      getBackendBaseUrl: () => string
      minimizeWindow: () => void
      toggleMaximizeWindow: () => void
      closeWindow: () => void
      showBubble: (text: string | null, emotion: string, audio?: string) => void
      sendConnectionAlive: () => void
      setClickThrough: (enabled: boolean) => void
      onBubbleShow: (callback: (data: { text: string | null; emotion?: string; audio?: string }) => void) => void
      onConnectionAlive: (callback: () => void) => void
    }
  }
}
