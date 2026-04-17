"use client"

import { useEffect } from "react"

import { OnboardingWizard } from "@/components/onboarding-wizard"
import { WindowControls } from "@/components/window-controls"

export default function OnboardingPage() {
  useEffect(() => {
    document.title = "初始化向导 - PetClaw AI"
  }, [])

  return (
    <div className="h-screen bg-background p-4 md:p-5">
      <div className="flex h-full flex-col overflow-hidden rounded-3xl border border-border/70 bg-card shadow-[0_20px_60px_-35px_rgba(120,80,20,0.45)]">
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
