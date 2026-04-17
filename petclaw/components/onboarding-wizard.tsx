"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import {
  AlarmClock,
  BookOpen,
  CalendarDays,
  Cat,
  Check,
  ChevronDown,
  FolderUp,
  Gauge,
  Loader2,
  Minus,
  MoonStar,
  PawPrint,
  Sparkles,
  Square,
  X,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Progress } from "@/components/ui/progress"
import { API_ENDPOINTS, getApiBaseUrl } from "@/lib/api"
import { fetchWithAuthRetry } from "@/lib/api/auth-bootstrap"
import { saveOnboardingState } from "@/lib/onboarding"
import { buildLearningRhythm, buildPressurePlan } from "@/lib/student-insights"

type SetupStatus = "pending" | "running" | "done" | "failed"
type PersonalityTone = "阴阳怪气" | "抽象发疯" | "甜心夹子"

interface SetupTask {
  id: string
  label: string
  detail: string
  status: SetupStatus
}

interface OnboardingWizardProps {
  onFinish: () => void
}

const steps = ["契约确认", "学习画像", "桌宠人格"]
const milestones = ["认识你", "理解你", "陪伴你"]

const setupTemplate: SetupTask[] = [
  { id: "env", label: "环境检查", detail: "检查 Launcher 服务连通性", status: "pending" },
  { id: "gateway", label: "网关状态", detail: "确认网关可用", status: "pending" },
  { id: "pico", label: "配置 Pet Channel", detail: "初始化桌宠通信通道", status: "pending" },
  { id: "token", label: "连接验证", detail: "拉取 Pico Token 并完成连通", status: "pending" },
]

const permissionLabels = {
  screenView: "屏幕学习状态感知",
  appCheck: "应用专注状态感知",
  localDocRead: "本地学习资料索引",
  popupReminder: "强提醒弹窗",
}

const permissionHints = {
  screenView: "仅用于判断你是否在学习场景，便于桌宠切换提醒语气。",
  appCheck: "检测是否切到娱乐应用，触发更有梗的拉回提醒。",
  localDocRead: "建立资料关键词索引，提问时能更快帮你定位文件。",
  popupReminder: "临近 DDL 时允许桌宠主动打断提醒，防止错过。",
}

const identityOptions = [
  { id: "student", label: "学生（已支持）", description: "当前版本完整支持学生场景引导。", enabled: true },
  { id: "teacher", label: "教师（开发中）", description: "后续会支持课程管理与授课节奏。", enabled: false },
  { id: "worker", label: "职场用户（开发中）", description: "后续会支持会议与任务优先级管理。", enabled: false },
] as const

const presetMajors = ["计算机", "法学", "临床医学", "汉语言", "金融", "艺术设计"]
const breakerVocabulary = ["高数", "四六级", "早八", "实验报告", "课程设计", "论文", "答辩"]
const nicknameTags = ["义父", "铲屎官", "大冤种", "少侠", "学术巨佬"]
const loadingComfortLines = [
  "我在替你检查连接，马上就好。",
  "正在和网关打招呼，别着急喵。",
  "快完成了，我会把环境整理干净。",
]

function withStatus(tasks: SetupTask[], id: string, status: SetupStatus): SetupTask[] {
  return tasks.map((task) => (task.id === id ? { ...task, status } : task))
}

function TaskIcon({ status }: { status: SetupStatus }) {
  if (status === "done") return <Check className="h-4 w-4 text-emerald-600" />
  if (status === "running") return <Loader2 className="h-4 w-4 animate-spin text-foreground" />
  if (status === "failed") return <span className="text-sm font-semibold text-red-600">!</span>
  return <span className="h-3 w-3 rounded-full border border-muted-foreground/50" />
}

export function OnboardingWizard({ onFinish }: OnboardingWizardProps) {
  const [step, setStep] = useState(0)
  const [userIdentity, setUserIdentity] = useState<"student" | "teacher" | "worker" | null>(null)

  const [setupTasks, setSetupTasks] = useState<SetupTask[]>(setupTemplate)
  const [setupRunning, setSetupRunning] = useState(false)
  const [setupError, setSetupError] = useState<string | null>(null)
  const [setupReady, setSetupReady] = useState(false)
  const [autoSetupStarted, setAutoSetupStarted] = useState(false)
  const [loadingLineIndex, setLoadingLineIndex] = useState(0)

  const [profile, setProfile] = useState({ displayName: "", role: "计算机", language: "中文" })
  const [app, setApp] = useState({
    autoConnectOnLaunch: true,
    enableDesktopBubble: true,
    openConsoleOnPetClick: true,
  })

  const [apiKey, setApiKey] = useState("")
  const [permissions, setPermissions] = useState({
    screenView: true,
    appCheck: true,
    localDocRead: true,
    popupReminder: true,
  })

  const [deadlineDate, setDeadlineDate] = useState("")
  const [anxietyLevel, setAnxietyLevel] = useState(58)
  const [sleepHour, setSleepHour] = useState(1)
  const [selectedBreakers, setSelectedBreakers] = useState<string[]>(["早八", "论文"])
  const [icsFileName, setIcsFileName] = useState("")
  const [isDropActive, setIsDropActive] = useState(false)

  const [personalityTone, setPersonalityTone] = useState<PersonalityTone>("阴阳怪气")
  const [activityLevel, setActivityLevel] = useState(65)
  const [nickname, setNickname] = useState("义父")
  const [customNickname, setCustomNickname] = useState("")
  const [showCompletionCard, setShowCompletionCard] = useState(false)
  const [previewReminderText, setPreviewReminderText] = useState("")

  const dateInputRef = useRef<HTMLInputElement | null>(null)
  const hasElectronApi = typeof window !== "undefined" && Boolean(window.electronAPI)

  const setupProgress = useMemo(() => {
    const done = setupTasks.filter((t) => t.status === "done").length
    return (done / setupTasks.length) * 100
  }, [setupTasks])

  const learningRhythmInsight = useMemo(() => {
    return buildLearningRhythm({
      major: profile.role,
      sleepHour,
      anxietyLevel,
      deadlineDate,
      selectedBreakers,
      hasScheduleFile: Boolean(icsFileName),
    })
  }, [anxietyLevel, deadlineDate, icsFileName, profile.role, selectedBreakers, sleepHour])

  const pressurePlanInsight = useMemo(() => {
    return buildPressurePlan({
      major: profile.role,
      sleepHour,
      anxietyLevel,
      deadlineDate,
      selectedBreakers,
      hasScheduleFile: Boolean(icsFileName),
    })
  }, [anxietyLevel, deadlineDate, icsFileName, profile.role, selectedBreakers, sleepHour])

  const moodTheme = useMemo(() => {
    if (pressurePlanInsight.level === 'critical' || pressurePlanInsight.level === 'high') {
      return {
        shellGradient: pressurePlanInsight.level === 'critical'
          ? "from-[#ffe3e6] via-[#ffd3d8] to-[#ffc2cb]"
          : "from-[#ffe7d7] via-[#ffd8bf] to-[#ffc79f]",
        sidePanel: pressurePlanInsight.level === 'critical' ? "bg-[#ffd0d7]" : "bg-[#ffd4b4]",
        accent: pressurePlanInsight.level === 'critical' ? "bg-[#ff5b5b] hover:bg-[#e54848]" : "bg-[#ff8a3d] hover:bg-[#f07324]",
        shellTint: pressurePlanInsight.level === 'critical' ? '#ffb5c3' : '#ffbf8b',
        sideTint: pressurePlanInsight.level === 'critical' ? '#ff9eb1' : '#ffb172',
      }
    }
    if (pressurePlanInsight.level === 'medium') {
      return {
        shellGradient: "from-[#fff2df] via-[#ffe9d1] to-[#fbdcbc]",
        sidePanel: "bg-[#ffe7cc]",
        accent: "bg-[#d66a12] hover:bg-[#bf5c0c]",
        shellTint: '#ffd79f',
        sideTint: '#ffc681',
      }
    }
    return {
      shellGradient: "from-[#e9fff4] via-[#f2fff8] to-[#dbfff0]",
      sidePanel: "bg-[#dbfaea]",
      accent: "bg-[#1f9f6f] hover:bg-[#16875c]",
      shellTint: '#bff4d8',
      sideTint: '#a2eec8',
    }
  }, [pressurePlanInsight.level])

  const companionFeedback = useMemo(() => {
    if (setupRunning) {
      return loadingComfortLines[loadingLineIndex]
    }
    if (setupError) {
      return `这次没连上，不是你的问题。${setupError}，我们可以再试一次。`
    }
    if (step === 0) {
      if (userIdentity !== "student") {
        return "先确认你是学生身份，我才能启用这套学习陪伴模型。"
      }
      if (setupReady) {
        return "准备完成，我已经能理解你的基础学习场景了。"
      }
      return "我在预热中，先把你的身份和偏好告诉我吧。"
    }
    if (step === 1) {
      if (icsFileName) {
        return `收到课表 ${icsFileName}，我会尽量避开上课时段提醒你。`
      }
      if (selectedBreakers.length > 3) {
        return "我看到你的压力点比较密集，后续提醒会更聚焦、少废话。"
      }
      return `已记录你的${profile.role}学习画像，我会按${sleepHour.toString().padStart(2, "0")}:00的作息来安排提醒。`
    }
    const calling = customNickname.trim() || nickname
    return `明白了，以后我会叫你“${calling}”，并按${personalityTone}风格陪你推进任务。`
  }, [customNickname, icsFileName, loadingLineIndex, nickname, personalityTone, profile.role, selectedBreakers.length, setupError, setupReady, setupRunning, sleepHour, step, userIdentity])

  const runSetup = async (): Promise<boolean> => {
    setSetupRunning(true)
    setSetupError(null)
    setSetupReady(false)
    setSetupTasks(setupTemplate)

    try {
      const baseUrl = getApiBaseUrl()

      setSetupTasks((prev) => withStatus(prev, "env", "running"))
      const authRes = await fetchWithAuthRetry(`${baseUrl}${API_ENDPOINTS.AUTH.STATUS}`, { credentials: "include" }, false)
      if (![200, 401].includes(authRes.status)) throw new Error(`环境检查失败（${authRes.status}）`)

      setSetupTasks((prev) => withStatus(withStatus(prev, "env", "done"), "gateway", "running"))
      const gatewayRes = await fetchWithAuthRetry(`${baseUrl}${API_ENDPOINTS.GATEWAY.STATUS}`, { credentials: "include" })
      if (![200, 401].includes(gatewayRes.status)) throw new Error(`网关检查失败（${gatewayRes.status}）`)

      setSetupTasks((prev) => withStatus(withStatus(prev, "gateway", "done"), "pico", "running"))
      let setupRes = await fetchWithAuthRetry(`${baseUrl}${API_ENDPOINTS.PET.SETUP}`, { method: "POST", credentials: "include" })
      if (setupRes.status === 404) {
        setupRes = await fetchWithAuthRetry(`${baseUrl}${API_ENDPOINTS.PICO.SETUP}`, { method: "POST", credentials: "include" })
      }
      if (!setupRes.ok) throw new Error(`Pet Channel 初始化失败（${setupRes.status}）`)

      setSetupTasks((prev) => withStatus(withStatus(prev, "pico", "done"), "token", "running"))
      let tokenRes = await fetchWithAuthRetry(`${baseUrl}${API_ENDPOINTS.PET.TOKEN}`, { credentials: "include" })
      if (tokenRes.status === 404) {
        tokenRes = await fetchWithAuthRetry(`${baseUrl}${API_ENDPOINTS.PICO.TOKEN}`, { credentials: "include" })
      }
      if (!tokenRes.ok) throw new Error(`连接验证失败（${tokenRes.status}）`)

      setSetupTasks((prev) => withStatus(prev, "token", "done"))
      setSetupReady(true)
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : "初始化失败"
      setSetupError(message)
      setSetupTasks((prev) => prev.map((task) => (task.status === "running" ? { ...task, status: "failed" } : task)))
      return false
    } finally {
      setSetupRunning(false)
    }
  }

  useEffect(() => {
    if (step !== 0 || autoSetupStarted) {
      return
    }
    setAutoSetupStarted(true)
    runSetup().catch(() => {})
  }, [autoSetupStarted, step])

  useEffect(() => {
    if (!setupRunning) {
      setLoadingLineIndex(0)
      return
    }
    const timer = window.setInterval(() => {
      setLoadingLineIndex((prev) => (prev + 1) % loadingComfortLines.length)
    }, 1800)

    return () => window.clearInterval(timer)
  }, [setupRunning])

  useEffect(() => {
    setShowCompletionCard(false)
  }, [step])

  useEffect(() => {
    if (typeof document === "undefined") {
      return
    }
    document.documentElement.setAttribute("data-mood-level", pressurePlanInsight.level)
  }, [pressurePlanInsight.level])

  const toggleBreaker = (value: string) => {
    setSelectedBreakers((prev) => (prev.includes(value) ? prev.filter((item) => item !== value) : [...prev, value]))
  }

  const pickIcsFile = (file: File | null) => {
    if (!file) return
    if (!file.name.toLowerCase().endsWith(".ics")) {
      setSetupError("只支持 .ics 课表文件，请重新选择")
      return
    }
    setSetupError(null)
    setIcsFileName(file.name)
  }

  const quickPickDeadline = () => {
    if (dateInputRef.current && typeof dateInputRef.current.showPicker === "function") {
      dateInputRef.current.showPicker()
      return
    }
    if (!deadlineDate) {
      const nextWeek = new Date()
      nextWeek.setDate(nextWeek.getDate() + 7)
      setDeadlineDate(nextWeek.toISOString().slice(0, 10))
    }
  }

  const previewReminder = () => {
    const candidate = pressurePlanInsight.level === 'critical'
      ? pressurePlanInsight.templates.strong
      : pressurePlanInsight.level === 'high'
        ? pressurePlanInsight.templates.normal
        : pressurePlanInsight.templates.soft

    setPreviewReminderText(candidate)
    if (window.electronAPI?.showBubble) {
      window.electronAPI.showBubble(candidate, 'neutral')
    }
  }

  const complete = async () => {
    let ready = setupReady
    if (!ready) {
      ready = await runSetup()
    }
    if (!ready) {
      setSetupError("当前连接初始化未完成，已先为你进入控制台；你可以稍后在配置页继续重试。")
    }

    const finalNickname = customNickname.trim() || nickname
    const voiceStyle = activityLevel >= 65 ? "活泼碎碎念" : "温和陪伴型"

    if (apiKey.trim()) {
      try {
        window.localStorage.setItem("petclaw.apiKeyDraft", apiKey.trim())
      } catch {
      }
    }

    try {
      window.localStorage.setItem("petclaw.userIdentity", userIdentity ?? "student")
    } catch {
    }

    try {
      window.localStorage.setItem(
        "petclaw.studentInsights",
        JSON.stringify({
          learningRhythm: learningRhythmInsight,
          pressurePlan: pressurePlanInsight,
        }),
      )
      window.localStorage.setItem("petclaw.moodLevel", pressurePlanInsight.level)
    } catch {
    }

    saveOnboardingState({
      version: 1,
      completed: true,
      completedAt: new Date().toISOString(),
      profile: {
        displayName: finalNickname,
        role: profile.role,
        language: profile.language,
      },
      pet: {
        petName: finalNickname,
        personality: personalityTone,
        voiceStyle,
      },
      app: {
        ...app,
        enableDesktopBubble: permissions.popupReminder,
      },
      studentInsights: {
        learningRhythm: learningRhythmInsight,
        pressurePlan: pressurePlanInsight,
      },
    })
    onFinish()
  }

  return (
    <div
      className={`relative h-full w-full overflow-hidden rounded-[2rem] border border-[#e8dccd] bg-gradient-to-br ${moodTheme.shellGradient} shadow-[0_30px_90px_-50px_rgba(130,84,23,0.5)] transition-[background-color,box-shadow] duration-700 ease-out`}
      style={{ backgroundColor: moodTheme.shellTint, backgroundBlendMode: 'multiply' }}
    >
      <div className="grid h-full min-h-0 grid-cols-1 md:grid-cols-[1.25fr_0.75fr]">
        <div className="flex h-full min-h-0 flex-col">
          <div className="border-b border-[#eadfce] px-6 py-5 md:px-8">
            <div className="mb-3 flex items-start justify-between gap-4">
              <div className="flex items-center gap-2">
                {steps.map((_, index) => (
                  <span key={index} className={`h-1.5 rounded-full transition-all ${index <= step ? "w-11 bg-[#9c5f22]" : "w-9 bg-[#e2d6c6]"}`} />
                ))}
              </div>
              <div className="flex items-center gap-1 rounded-xl border border-[#e4d4c0] bg-white/80 p-1 shadow-sm">
                <button
                  type="button"
                  className="rounded-md p-1.5 text-[#7a6755] hover:bg-[#f6ebdd] disabled:cursor-not-allowed disabled:opacity-40"
                  onClick={() => window.electronAPI?.minimizeWindow?.()}
                  disabled={!hasElectronApi}
                  title="最小化"
                >
                  <Minus className="h-3.5 w-3.5" />
                </button>
                <button
                  type="button"
                  className="rounded-md p-1.5 text-[#7a6755] hover:bg-[#f6ebdd] disabled:cursor-not-allowed disabled:opacity-40"
                  onClick={() => window.electronAPI?.toggleMaximizeWindow?.()}
                  disabled={!hasElectronApi}
                  title="最大化 / 还原"
                >
                  <Square className="h-3.5 w-3.5" />
                </button>
                <button
                  type="button"
                  className="rounded-md p-1.5 text-[#7a6755] hover:bg-[#f8d9d6] hover:text-[#b42318]"
                  onClick={() => {
                    if (hasElectronApi) {
                      window.electronAPI?.closeWindow?.()
                    } else {
                      window.close()
                    }
                  }}
                  title="关闭"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>
            </div>
            <h1 className="text-2xl font-semibold text-[#3f2d1f] md:text-3xl">{steps[step]}</h1>
            <p className="mt-1 text-sm text-[#7a6755]">阶段：{milestones[step]} · 已完成 {step + 1}/3</p>
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto px-6 py-6 md:px-8">
            {step === 0 && (
              <div className="space-y-5">
                <div className="rounded-2xl border border-[#ecd8bd] bg-[#fff4e5] p-5">
                  <div className="mb-2 flex items-center gap-2 text-xl font-semibold text-[#3f2d1f]">
                    <Sparkles className="h-5 w-5 text-[#d4812f]" />
                    注入灵力源（API Key）
                  </div>
                  <Input
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    placeholder="可选：填写你自己的 API Key，不填也可继续"
                    className="mt-3 rounded-full border-[#e8cfae] bg-white"
                  />
                  <p className="mt-2 text-xs text-[#7a6755]">用于切换到你的私有模型额度，不会影响默认可用功能。</p>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-2 text-lg font-semibold text-[#3f2d1f]">你目前的身份</p>
                  <p className="mb-3 text-sm text-[#7a6755]">当前初始化流程专为学生设计，请先选择“学生”继续，其他身份将后续开放。</p>
                  <div className="grid gap-2">
                    {identityOptions.map((item) => {
                      const selected = userIdentity === item.id
                      return (
                        <button
                          key={item.id}
                          type="button"
                          disabled={!item.enabled}
                          onClick={() => setUserIdentity(item.id)}
                          className={`rounded-xl border px-4 py-3 text-left transition ${selected ? "border-[#d17a2a] bg-[#ffe8cc]" : "border-[#e2d4c3] bg-white"} ${item.enabled ? "hover:border-[#d17a2a]" : "cursor-not-allowed opacity-60"}`}
                        >
                          <p className="text-sm font-medium text-[#4d3929]">{item.label}</p>
                          <p className="text-xs text-[#7a6755]">{item.description}</p>
                        </button>
                      )
                    })}
                  </div>
                </div>

                <div className="space-y-2 rounded-2xl border border-dashed border-[#e6cfb0] bg-[#fff7eb] p-5">
                  <p className="text-sm font-medium text-[#5a4733]">为什么要这些权限？为了让桌宠更懂你的学习状态，而不是“乱提醒”。</p>
                  {(Object.keys(permissions) as Array<keyof typeof permissions>).map((key) => (
                    <label key={key} className="flex cursor-pointer items-start gap-3 rounded-lg px-1 py-1 text-sm text-[#4d3929]">
                      <input
                        type="checkbox"
                        checked={permissions[key]}
                        onChange={(e) => setPermissions((prev) => ({ ...prev, [key]: e.target.checked }))}
                        className="mt-1 h-4 w-4 accent-[#d17a2a]"
                      />
                      <span>
                        <span className="block font-medium">{permissionLabels[key]}</span>
                        <span className="block text-xs text-[#7a6755]">{permissionHints[key]}</span>
                      </span>
                    </label>
                  ))}
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-4">
                  <div className="mb-3 flex items-center justify-between text-sm text-[#5a4733]">
                    <span className="font-medium">连接体检进度</span>
                    <span>{Math.round(setupProgress)}%</span>
                  </div>
                  <Progress value={setupProgress} className="h-2 [&>div]:bg-[#cd7a2f]" />
                  <p className="mt-2 text-xs text-[#7a6755]">
                    {setupRunning ? loadingComfortLines[loadingLineIndex] : setupReady ? "环境准备完成，可以继续完善学习画像。" : "首次进入会自动体检，失败可点击“重新体检”。"}
                  </p>
                  <div className="mt-3 grid gap-2">
                    {setupTasks.map((task) => (
                      <div key={task.id} className="flex items-start gap-3 rounded-xl border border-[#eadfce] bg-white px-3 py-2">
                        <div className="mt-0.5 h-4 w-4"><TaskIcon status={task.status} /></div>
                        <div>
                          <p className="text-sm font-medium text-[#3f2d1f]">{task.label}</p>
                          <p className="text-xs text-[#7a6755]">{task.detail}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {setupError && <div className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{setupError}</div>}

                <div className="flex flex-wrap gap-3">
                  <Button onClick={() => runSetup()} disabled={setupRunning} className={`min-w-44 text-white transition-colors duration-500 ${moodTheme.accent}`}>
                    {setupRunning ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PawPrint className="mr-2 h-4 w-4" />}重新体检
                  </Button>
                </div>
              </div>
            )}

            {step === 1 && (
              <div className="space-y-5">
                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-4 text-sm text-[#6a5644]">
                  这些信息会用于生成你的学习画像：桌宠会基于你的作息、课程压力和高频痛点来调整提醒时机与语气。
                </div>
                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-3 flex items-center gap-2 text-2xl font-semibold text-[#3f2d1f]"><BookOpen className="h-5 w-5 text-[#d17a2a]" />你的主修功法（专业）</p>
                  <div className="relative">
                    <select
                      value={profile.role}
                      onChange={(e) => setProfile((prev) => ({ ...prev, role: e.target.value }))}
                      className="w-full appearance-none rounded-full border border-[#e4d4c0] bg-white px-4 py-2.5 text-lg text-[#4d3929]"
                    >
                      {presetMajors.map((major) => (
                        <option key={major} value={major}>{major}</option>
                      ))}
                    </select>
                    <ChevronDown className="pointer-events-none absolute right-4 top-1/2 h-4 w-4 -translate-y-1/2 text-[#7a6755]" />
                  </div>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-3 flex items-center gap-2 text-2xl font-semibold text-[#3f2d1f]"><AlarmClock className="h-5 w-5 text-[#d17a2a]" />最近的渡劫日（下个重要 DDL）</p>
                  <div className="flex items-center gap-2">
                    <Input
                      ref={dateInputRef}
                      type="date"
                      value={deadlineDate}
                      onChange={(e) => setDeadlineDate(e.target.value)}
                      className="rounded-full border-[#e4d4c0] text-[#4d3929]"
                    />
                    <button
                      type="button"
                      onClick={quickPickDeadline}
                      className="rounded-full border border-[#e4d4c0] bg-white p-2.5 text-[#d17a2a] hover:bg-[#fff3e3]"
                      title="选择日期"
                    >
                      <CalendarDays className="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-3 flex items-center gap-2 text-2xl font-semibold text-[#3f2d1f]"><MoonStar className="h-5 w-5 text-[#d17a2a]" />作息时间偏好</p>
                  <p className="mb-2 text-sm text-[#6a5644]">常见入睡时间：{sleepHour.toString().padStart(2, "0")}:00</p>
                  <input
                    type="range"
                    min={0}
                    max={23}
                    value={sleepHour}
                    onChange={(e) => setSleepHour(Number(e.target.value))}
                    className="h-2 w-full cursor-pointer appearance-none rounded-full bg-gradient-to-r from-[#f4dcb8] to-[#d6c3a8]"
                  />
                  <div className="mt-2 flex justify-between text-xs text-[#7a6755]">
                    <span>23:00 前（养生）</span>
                    <span>03:00+（修仙）</span>
                  </div>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-3 text-2xl font-semibold text-[#3f2d1f]">破防词汇（多选）</p>
                  <p className="mb-2 text-sm text-[#7a6755]">选你最容易焦虑的关键词，桌宠会在相关任务中提高提醒频次。</p>
                  <div className="flex flex-wrap gap-2">
                    {breakerVocabulary.map((item) => {
                      const active = selectedBreakers.includes(item)
                      return (
                        <button
                          key={item}
                          type="button"
                          onClick={() => toggleBreaker(item)}
                          className={`rounded-full border px-4 py-1.5 text-sm transition ${active ? "border-[#d17a2a] bg-[#ffe8cc] text-[#7b4a16]" : "border-[#dac7b1] bg-white text-[#6a5644]"}`}
                        >
                          [{item}]
                        </button>
                      )
                    })}
                  </div>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-3 text-2xl font-semibold text-[#3f2d1f]">课表导入（.ics）</p>
                  <p className="mb-2 text-sm text-[#7a6755]">导入后可自动识别上课时段，避免在上课时高频打扰。</p>
                  <label
                    onDragOver={(e) => {
                      e.preventDefault()
                      setIsDropActive(true)
                    }}
                    onDragLeave={() => setIsDropActive(false)}
                    onDrop={(e) => {
                      e.preventDefault()
                      setIsDropActive(false)
                      pickIcsFile(e.dataTransfer.files?.[0] ?? null)
                    }}
                    className={`block cursor-pointer rounded-2xl border-2 border-dashed p-8 text-center transition ${isDropActive ? "border-[#d17a2a] bg-[#fff0dc]" : "border-[#d8c6ae] bg-white"}`}
                  >
                    <FolderUp className="mx-auto mb-3 h-10 w-10 text-[#b8742f]" />
                    <p className="text-xl font-semibold text-[#4d3929]">拖拽 .ics 课表文件到这里</p>
                    <p className="mt-2 text-sm text-[#7a6755]">或点击选择文件</p>
                    <input
                      type="file"
                      accept=".ics"
                      className="hidden"
                      onChange={(e) => pickIcsFile(e.target.files?.[0] ?? null)}
                    />
                  </label>
                  {icsFileName ? <p className="mt-2 text-sm text-emerald-700">已接收：{icsFileName}</p> : <p className="mt-2 text-sm text-[#7a6755]">可选：没有课表也能继续</p>}
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-3 text-2xl font-semibold text-[#3f2d1f]">学习压力感知</p>
                  <input
                    type="range"
                    min={0}
                    max={100}
                    value={anxietyLevel}
                    onChange={(e) => setAnxietyLevel(Number(e.target.value))}
                    className="h-2 w-full cursor-pointer appearance-none rounded-full bg-gradient-to-r from-[#f2b8bd] via-[#f3dfb3] to-[#c5d2f8]"
                  />
                  <div className="mt-2 flex justify-between text-sm text-[#7a6755]">
                    <span>心态平稳</span>
                    <span>快碎了</span>
                  </div>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-2 text-xl font-semibold text-[#3f2d1f]">将写入桌宠的学习节奏画像</p>
                  <p className="mb-3 text-sm text-[#7a6755]">用于方案 1：让提醒时间更像你的真实学习节奏。</p>
                  <div className="rounded-xl border border-[#e4d4c0] bg-white p-3 text-sm text-[#5a4733]">
                    <p className="font-medium">{learningRhythmInsight.summary}</p>
                    <p className="mt-2">高专注窗口：{learningRhythmInsight.focusWindows.join(" / ")}</p>
                    <p className="mt-1">静默策略：{learningRhythmInsight.quietWindows.join("；")}</p>
                  </div>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-2 text-xl font-semibold text-[#3f2d1f]">将写入桌宠的压力分层策略</p>
                  <p className="mb-3 text-sm text-[#7a6755]">用于方案 2：按压力等级切换提醒频率与话术强度。</p>
                  <div className="space-y-2 rounded-xl border border-[#e4d4c0] bg-white p-3 text-sm text-[#5a4733]">
                    <p>压力等级：<span className="font-semibold">{pressurePlanInsight.level}</span></p>
                    <p>提醒间隔（分钟）：{pressurePlanInsight.reminderIntervalsMinutes.join(" / ")}</p>
                    <p>策略：{pressurePlanInsight.strategy}</p>
                    <p>话术风格：{pressurePlanInsight.toneGuide}</p>
                  </div>
                </div>
              </div>
            )}

            {step === 2 && (
              <div className="space-y-5">
                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-3 flex items-center gap-2 text-2xl font-semibold text-[#3f2d1f]"><Cat className="h-5 w-5 text-[#d17a2a]" />嘴替流派（Personality）</p>
                  <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                    {(["阴阳怪气", "抽象发疯", "甜心夹子"] as PersonalityTone[]).map((tone) => (
                      <button
                        key={tone}
                        type="button"
                        onClick={() => setPersonalityTone(tone)}
                        className={`rounded-xl border px-3 py-2 text-sm transition ${personalityTone === tone ? "border-[#d17a2a] bg-[#ffe8cc] text-[#7b4a16]" : "border-[#d8c6ae] bg-white text-[#6a5644]"}`}
                      >
                        {tone}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-3 flex items-center gap-2 text-2xl font-semibold text-[#3f2d1f]"><Gauge className="h-5 w-5 text-[#d17a2a]" />显眼包指数（Activity Level）</p>
                  <input
                    type="range"
                    min={0}
                    max={100}
                    value={activityLevel}
                    onChange={(e) => setActivityLevel(Number(e.target.value))}
                    className="h-2 w-full cursor-pointer appearance-none rounded-full bg-gradient-to-r from-[#f5c58a] to-[#de8f3f]"
                  />
                  <div className="mt-2 flex justify-between text-sm text-[#7a6755]">
                    <span>高冷自闭</span>
                    <span>满屏乱爬</span>
                  </div>
                </div>

                <div className="rounded-2xl border border-[#eadfce] bg-[#fffaf4] p-5">
                  <p className="mb-1 text-2xl font-semibold text-[#3f2d1f]">桌宠怎么称呼你</p>
                  <p className="mb-3 text-sm text-[#7a6755]">这是桌宠对你的称呼，不是你的网名。会影响它的对话语气和亲近感。</p>
                  <div className="mb-3 flex flex-wrap gap-2">
                    {nicknameTags.map((tag) => (
                      <button
                        key={tag}
                        type="button"
                        onClick={() => setNickname(tag)}
                        className={`rounded-full border px-3 py-1.5 text-sm ${nickname === tag ? "border-[#d17a2a] bg-[#ffe8cc] text-[#7b4a16]" : "border-[#d8c6ae] bg-white text-[#6a5644]"}`}
                      >
                        [{tag}]
                      </button>
                    ))}
                  </div>
                  <Input
                    value={customNickname}
                    onChange={(e) => {
                      setCustomNickname(e.target.value)
                      setProfile((prev) => ({ ...prev, displayName: e.target.value }))
                    }}
                    placeholder="或者自定义称呼"
                    className="border-[#dac7b1] bg-white text-[#4d3929]"
                  />
                </div>

                <label className="flex items-center justify-between rounded-xl border border-[#d8c6ae] bg-[#fffaf4] p-3 text-sm text-[#5a4733]">
                  启动后自动连接
                  <input
                    type="checkbox"
                    checked={app.autoConnectOnLaunch}
                    onChange={(e) => setApp((prev) => ({ ...prev, autoConnectOnLaunch: e.target.checked }))}
                    className="h-4 w-4 accent-[#d17a2a]"
                  />
                </label>

                <label className="flex items-center justify-between rounded-xl border border-[#d8c6ae] bg-[#fffaf4] p-3 text-sm text-[#5a4733]">
                  点击桌宠打开控制台
                  <input
                    type="checkbox"
                    checked={app.openConsoleOnPetClick}
                    onChange={(e) => setApp((prev) => ({ ...prev, openConsoleOnPetClick: e.target.checked }))}
                    className="h-4 w-4 accent-[#d17a2a]"
                  />
                </label>
              </div>
            )}
          </div>

          <div className="shrink-0 border-t border-[#eadfce] bg-[#fff8ef]/95 px-6 py-5 backdrop-blur md:px-8">
            <div className="flex items-center justify-between">
            <Button variant="ghost" onClick={() => setStep((prev) => Math.max(0, prev - 1))} disabled={step === 0 || setupRunning} className="text-[#6a5644]">
              上一步
            </Button>
            {step < 2 ? (
              <Button
                onClick={() => setStep((prev) => Math.min(2, prev + 1))}
                disabled={setupRunning || (step === 0 && userIdentity !== "student")}
                className={`rounded-full px-6 text-white transition-colors duration-500 ${moodTheme.accent}`}
              >
                下一步喵~
              </Button>
            ) : (
              <Button onClick={() => setShowCompletionCard(true)} disabled={setupRunning} className="rounded-full bg-[#ff9100] px-6 text-white hover:bg-[#e67f00]">
                {setupRunning ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Check className="mr-2 h-4 w-4" />}大功告成，召唤桌宠
              </Button>
            )}
            </div>
            {step === 0 && userIdentity !== "student" && (
              <p className="mt-2 text-right text-xs text-[#8a735c]">请先选择“学生（已支持）”后继续。</p>
            )}
          </div>
        </div>

        <div
          className={`hidden h-full border-l border-[#eadfce] ${moodTheme.sidePanel} p-8 transition-colors duration-700 ease-out md:block`}
          style={{ backgroundColor: moodTheme.sideTint }}
        >
          <div className="mb-6 flex items-center gap-2 rounded-full bg-white/80 px-3 py-1 text-xs text-[#6a5644] backdrop-blur">
            <Sparkles className="h-3.5 w-3.5" />PetClaw 初始化向导
          </div>
          <div className="rounded-2xl border border-[#e8dccd] bg-white/90 px-4 py-3 text-sm leading-relaxed text-[#5a4733] shadow-sm">
            {companionFeedback}
          </div>
          <div className="mt-6 rounded-3xl border border-[#e8dccd] bg-white/80 p-6 shadow-sm">
            <div className="flex h-[20rem] flex-col items-center justify-center rounded-3xl bg-[#fff3e1] text-center">
              <span className="text-7xl">🐾</span>
              <p className="mt-4 max-w-xs text-sm text-[#6a5644]">完成三步解锁专属桌宠，学习提醒和情绪陪伴将根据你的节奏自动调节。</p>
            </div>
          </div>
        </div>
      </div>

      {showCompletionCard && (
        <div className="absolute inset-0 z-30 flex items-center justify-center bg-black/25 p-6">
          <div className="w-full max-w-2xl rounded-3xl border border-[#e8dccd] bg-[#fffaf4] p-6 shadow-xl">
            <p className="text-2xl font-semibold text-[#3f2d1f]">你的人设画像已准备好</p>
            <p className="mt-2 text-sm text-[#7a6755]">这一步完成后，桌宠将按你的学习节奏和压力层级进行提醒。</p>

            <div className="mt-4 grid gap-3 rounded-2xl border border-[#eadfce] bg-white p-4 text-sm text-[#5a4733]">
              <p><span className="font-medium">学习节奏：</span>{learningRhythmInsight.summary}</p>
              <p><span className="font-medium">高专注窗口：</span>{learningRhythmInsight.focusWindows.join(" / ")}</p>
              <p><span className="font-medium">压力等级：</span>{pressurePlanInsight.level}（{pressurePlanInsight.strategy}）</p>
              <p><span className="font-medium">提醒间隔：</span>{pressurePlanInsight.reminderIntervalsMinutes.join(" / ")} 分钟</p>
            </div>

            <div className="mt-4 rounded-2xl border border-dashed border-[#e4d4c0] bg-[#fff6eb] p-4">
              <p className="text-sm font-medium text-[#5a4733]">首条提醒预览</p>
              <p className="mt-1 text-sm text-[#6a5644]">{previewReminderText || "点击“试试第一条提醒”预览桌宠语气。"}</p>
            </div>

            <div className="mt-5 flex flex-wrap justify-end gap-2">
              <Button variant="outline" onClick={() => setShowCompletionCard(false)} className="border-[#e4d4c0] text-[#5a4733]">再看一下</Button>
              <Button variant="outline" onClick={previewReminder} className="border-[#e4d4c0] text-[#5a4733]">试试第一条提醒</Button>
              <Button onClick={complete} className="bg-[#ff9100] text-white hover:bg-[#e67f00]">确认并召唤桌宠</Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
