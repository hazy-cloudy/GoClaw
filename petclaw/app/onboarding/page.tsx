"use client"

import { useEffect, useState } from "react"

import { OnboardingWizard } from "@/components/onboarding-wizard"
import { WindowControls } from "@/components/window-controls"

export default function OnboardingPage() {
  const [isElectronShell, setIsElectronShell] = useState(false)

  useEffect(() => {
    setIsElectronShell(Boolean(window.electronAPI))
  }, [])

  useEffect(() => {
    document.title = "初始化向导 - PetClaw AI"
  }, [])

  return (
    <div className={`h-screen bg-background ${isElectronShell ? "p-0" : "p-4 md:p-5"}`}>
      <div className={`flex h-full flex-col overflow-hidden bg-card ${isElectronShell ? "rounded-none border-0 shadow-none" : "rounded-3xl border border-border/70 shadow-[0_20px_60px_-35px_rgba(120,80,20,0.45)]"}`}>
        <WindowControls title="初始化向导 - PetClaw AI" />
        <div className="min-h-0 flex-1">
          <OnboardingWizard
            onFinish={() => {
              if (typeof window !== "undefined" && window.opener) {
                window.close()
                return
              }
              window.location.href = "/"
            }}
          />
        </div>
      </div>
    </div>
  )
}
