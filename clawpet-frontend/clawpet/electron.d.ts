export {}

declare global {
  interface BubblePayload {
    text: string | null
    emotion: string
    level?: string
    animation?: string
    animationHints?: string[]
    audio?: string
    duration_ms?: number
  }

  interface PetRuntimeStatePayload {
    state:
      | "idle"
      | "listening"
      | "recognizing"
      | "thinking"
      | "tool_running"
      | "speaking"
      | "done"
      | "error"
      | "stalled"
    text?: string
    source?: "voice" | "chat"
    sticky_ms?: number
  }

  interface Window {
    electronAPI?: {
      openOnboarding?: () => void
      completeOnboarding?: () => void
      openSettings?: () => void
      setOnboardingMode?: (enabled: boolean) => void
      setPetClickThrough?: (enabled: boolean) => void
      getBackendBaseUrl?: () => string
      getApiBaseUrl?: () => string
      getLauncherToken?: () => string
      minimizeWindow?: () => void
      toggleMaximizeWindow?: () => void
      closeWindow?: () => void
      showBubble?: {
        (payload: BubblePayload): void
        (text: string | null, emotion: string, audio?: string): void
      }
      reportBubbleWindowSize?: (size: { width: number; height: number }) => void
      onSettingsUpdate?: (callback: (settings: unknown) => void) => void
      onBubbleShow?: (callback: (payload: BubblePayload) => void) => void
      onForceStopMedia?: (callback: () => void) => () => void
      ensureSettingsForeground?: () => Promise<{ ok: boolean; reason?: string }>
      onVoiceShortcutTriggered?: (callback: () => void) => void
      onVoiceInputStateChanged?: (
        callback: (payload: { phase?: string; isListening?: boolean }) => void,
      ) => (() => void) | void
      onPetRuntimeStateChanged?: (
        callback: (payload: PetRuntimeStatePayload) => void,
      ) => (() => void) | void
      registerVoiceShortcut?: (data: { enabled: boolean; keys: string }) => void
      registerPetClickThroughShortcut?: (data: { enabled: boolean; keys: string }) => void
      reportVoiceInputState?: (data: { phase?: string; isListening?: boolean }) => void
      reportPetRuntimeState?: (data: PetRuntimeStatePayload) => void
    }
  }
}
