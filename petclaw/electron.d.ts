export {}

declare global {
  interface BubblePayload {
    text: string | null
    emotion: string
    audio?: string
  }

  interface Window {
    electronAPI?: {
      openOnboarding?: () => void
      setOnboardingMode?: (enabled: boolean) => void
      getBackendBaseUrl?: () => string
      minimizeWindow?: () => void
      toggleMaximizeWindow?: () => void
      closeWindow?: () => void
      showBubble?: {
        (payload: BubblePayload): void
        (text: string | null, emotion: string, audio?: string): void
      }
    }
  }
}
