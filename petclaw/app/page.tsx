"use client"

import { useEffect, useState } from "react"
import dynamic from "next/dynamic"

import { Sidebar } from "@/components/sidebar"
import { ChatArea } from "@/components/chat-area"
import { WindowControls } from "@/components/window-controls"
import { loadOnboardingState } from "@/lib/onboarding"

const pageLoadingFallback = (
  <div className="flex h-full items-center justify-center text-sm text-muted-foreground">加载中...</div>
)

const OnboardingWizard = dynamic(
  () => import("@/components/onboarding-wizard").then((mod) => mod.OnboardingWizard),
  { loading: () => pageLoadingFallback }
)
const SkillsPage = dynamic(
  () => import("@/components/skills-page").then((mod) => mod.SkillsPage),
  { loading: () => pageLoadingFallback }
)
const SchedulePage = dynamic(
  () => import("@/components/schedule-page").then((mod) => mod.SchedulePage),
  { loading: () => pageLoadingFallback }
)
const ConfigPage = dynamic(
  () => import("@/components/config-page").then((mod) => mod.ConfigPage),
  { loading: () => pageLoadingFallback }
)

export default function Home() {
  const [activeNav, setActiveNav] = useState("聊天")
  const [bootState, setBootState] = useState<"checking" | "onboarding" | "ready">("ready")
  const [uiMood, setUiMood] = useState<"low" | "medium" | "high" | "critical">("medium")
  const [readyVisible, setReadyVisible] = useState(false)
  const [isElectronShell, setIsElectronShell] = useState(false)

  const applyMoodFromOnboardingState = () => {
    const state = loadOnboardingState()
    const level = state?.studentInsights?.pressurePlan?.level
    if (level === "low" || level === "medium" || level === "high" || level === "critical") {
      setUiMood(level)
      return
    }
    setUiMood("medium")
  }

  useEffect(() => {
    setIsElectronShell(Boolean(window.electronAPI))
  }, [])

  useEffect(() => {
    try {
      const hasForcedOnboarding = new URLSearchParams(window.location.search).get("onboarding") === "1"
      if (hasForcedOnboarding) {
        setBootState("onboarding")
        return
      }

      const onboardingState = loadOnboardingState()
      if (onboardingState?.completed) {
        applyMoodFromOnboardingState()
        setBootState("ready")
        return
      }
      setBootState("onboarding")
    } catch {
      setBootState("ready")
    }
  }, [])

  useEffect(() => {
    if (bootState === "onboarding") {
      document.title = "初始化向导 - PetClaw AI"
      return
    }

    if (bootState === "ready") {
      document.title = `${activeNav} - PetClaw AI`
    }
  }, [activeNav, bootState])

  useEffect(() => {
    if (!window.electronAPI?.setOnboardingMode) {
      return
    }

    const onboardingActive = bootState === "onboarding"
    window.electronAPI.setOnboardingMode(onboardingActive)

    return () => {
      window.electronAPI?.setOnboardingMode?.(false)
    }
  }, [bootState])

  useEffect(() => {
    if (typeof window === "undefined") {
      return
    }

    const root = document.documentElement
    const onboardingState = loadOnboardingState()
    const levelFromState = onboardingState?.studentInsights?.pressurePlan?.level
    const levelFromStorage = window.localStorage.getItem("petclaw.moodLevel") as "low" | "medium" | "high" | "critical" | null
    const nextLevel = levelFromState || levelFromStorage

    if (nextLevel) {
      root.setAttribute("data-mood-level", nextLevel)
    } else {
      root.removeAttribute("data-mood-level")
    }
  }, [bootState])

  useEffect(() => {
    if (bootState !== "ready") {
      setReadyVisible(false)
      return
    }

    const timer = window.setTimeout(() => {
      setReadyVisible(true)
    }, 30)

    return () => window.clearTimeout(timer)
  }, [bootState])

  const renderContent = () => {
    switch (activeNav) {
      case "聊天":
        return <ChatArea />
      case "技能":
        return <SkillsPage />
      case "定时任务":
        return <SchedulePage />
      case "配置":
        return <ConfigPage />
      default:
        return <ChatArea />
    }
  }

  if (bootState === "checking") {
    return (
      <div className="flex h-screen items-center justify-center bg-transparent text-sm text-muted-foreground">
        正在初始化...
      </div>
    )
  }

  if (bootState === "onboarding") {
    return (
      <div className={`h-screen bg-transparent ${isElectronShell ? "p-0" : "p-4 md:p-5"}`}>
        <OnboardingWizard
          onFinish={() => {
            const url = new URL(window.location.href)
            url.searchParams.delete("onboarding")
            window.history.replaceState({}, "", url.toString())
            applyMoodFromOnboardingState()
            setBootState("ready")
          }}
        />
      </div>
    )
  }

  const shellByMood = {
    low: "border-emerald-200/80 bg-gradient-to-br from-emerald-50/90 via-teal-50/80 to-white shadow-[0_20px_70px_-40px_rgba(16,185,129,0.5)]",
    medium: "border-amber-200/80 bg-gradient-to-br from-amber-50/90 via-orange-50/80 to-white shadow-[0_20px_70px_-40px_rgba(217,119,6,0.45)]",
    high: "border-orange-300/80 bg-gradient-to-br from-orange-50/90 via-amber-50/80 to-rose-50/50 shadow-[0_20px_70px_-35px_rgba(249,115,22,0.55)]",
    critical: "border-rose-300/90 bg-gradient-to-br from-rose-50/90 via-pink-50/80 to-orange-50/50 shadow-[0_20px_75px_-30px_rgba(244,63,94,0.6)]",
  } as const

  return (
    <div className={`h-screen bg-transparent transition-all duration-500 ease-out ${isElectronShell ? "p-0" : "p-4 md:p-5"} ${readyVisible ? "translate-y-0 opacity-100" : "translate-y-2 opacity-0"}`}>
      <div className={`flex h-full flex-col overflow-hidden transition-[background-color,border-color,box-shadow,border-radius] duration-700 ease-out ${isElectronShell ? "rounded-none border-0 shadow-none" : "rounded-3xl"} ${shellByMood[uiMood]}`}>
        <WindowControls title={`${activeNav} - PetClaw AI`} />
        <div className="flex min-h-0 flex-1" data-electron-no-drag="true">
          <Sidebar activeNav={activeNav} setActiveNav={setActiveNav} />
          {renderContent()}
        </div>
      </div>
    </div>
  )
}
