import {
  IconPlugConnectedX,
  IconRobot,
  IconRobotOff,
  IconStar,
  IconTool,
} from "@tabler/icons-react"
import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import type { ToolDataSource } from "@/api/tools"

interface ChatEmptyStateProps {
  hasAvailableModels: boolean
  defaultModelName: string
  isConnected: boolean
  enabledToolCount?: number
  toolInfoSource?: ToolDataSource
}

export function ChatEmptyState({
  hasAvailableModels,
  defaultModelName,
  isConnected,
  enabledToolCount,
  toolInfoSource,
}: ChatEmptyStateProps) {
  const { t } = useTranslation()

  if (!hasAvailableModels) {
    return (
      <div className="flex flex-col items-center justify-center py-20 opacity-70">
        <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-500">
          <IconRobotOff className="h-8 w-8" />
        </div>
        <h3 className="mb-2 text-xl font-medium">
          {t("chat.empty.noConfiguredModel")}
        </h3>
        <p className="text-muted-foreground mb-4 text-center text-sm">
          {t("chat.empty.noConfiguredModelDescription")}
        </p>
        <Button asChild variant="outline" size="sm" className="px-4">
          <Link to="/models">{t("chat.empty.goToModels")}</Link>
        </Button>
      </div>
    )
  }

  if (!defaultModelName) {
    return (
      <div className="flex flex-col items-center justify-center py-20 opacity-70">
        <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-500">
          <IconStar className="h-8 w-8" />
        </div>
        <h3 className="mb-2 text-xl font-medium">
          {t("chat.empty.noSelectedModel")}
        </h3>
        <p className="text-muted-foreground mb-4 text-center text-sm">
          {t("chat.empty.noSelectedModelDescription")}
        </p>
      </div>
    )
  }

  if (!isConnected) {
    return (
      <div className="flex flex-col items-center justify-center py-20 opacity-70">
        <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-500">
          <IconPlugConnectedX className="h-8 w-8" />
        </div>
        <h3 className="mb-2 text-xl font-medium">
          {t("chat.empty.notRunning")}
        </h3>
        <p className="text-muted-foreground mb-4 text-center text-sm">
          {t("chat.empty.notRunningDescription")}
        </p>
      </div>
    )
  }

  const showToolHint = enabledToolCount !== undefined && enabledToolCount <= 0

  return (
    <div className="flex flex-col items-center justify-center py-20 opacity-70">
      <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-violet-500/10 text-violet-500">
        <IconRobot className="h-8 w-8" />
      </div>
      <h3 className="mb-2 text-xl font-medium">{t("chat.welcome")}</h3>
      <p className="text-muted-foreground text-center text-sm">
        {t("chat.welcomeDesc")}
      </p>

      {showToolHint && (
        <div className="mt-5 rounded-xl border border-amber-300/60 bg-amber-50/60 px-4 py-3 text-center text-xs text-amber-800">
          <div className="mb-2 flex items-center justify-center gap-1.5 font-medium">
            <IconTool className="h-3.5 w-3.5" />
            <span>No tools are enabled</span>
          </div>
          <p>Enable at least one tool to allow ClawPet to run tool actions.</p>
          <Button asChild size="sm" variant="outline" className="mt-3 h-7 px-3">
            <Link to="/agent/tools">Open Tools</Link>
          </Button>
        </div>
      )}

      {toolInfoSource === "mock" && (
        <p className="text-muted-foreground mt-4 text-center text-xs">
          Tool status is currently from mock data.
        </p>
      )}
    </div>
  )
}
