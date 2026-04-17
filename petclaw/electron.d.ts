export {}

declare global {
  interface Window {
    electronAPI?: {
      openOnboarding?: () => void
      setOnboardingMode?: (enabled: boolean) => void
      getBackendBaseUrl?: () => string
      minimizeWindow?: () => void
      toggleMaximizeWindow?: () => void
      closeWindow?: () => void
      showBubble?: (text: string | null, emotion: string, audio?: string) => void
    }
  }
}
