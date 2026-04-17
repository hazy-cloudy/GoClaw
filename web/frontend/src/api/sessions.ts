import { launcherFetch } from "@/api/http"

export interface SessionSummary {
  id: string
  title: string
  preview: string
  message_count: number
  created: string
  updated: string
}

export interface SessionDetail {
  id: string
  messages: {
    role: "user" | "assistant"
    content: string
    media?: string[]
  }[]
  summary: string
  created: string
  updated: string
}

class ApiRequestError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = "ApiRequestError"
    this.status = status
  }
}

const ENABLE_MOCK_FALLBACK =
  import.meta.env.DEV || import.meta.env.VITE_ENABLE_MOCK_DATA === "1"

const FALLBACK_STATUS_CODES = new Set([404, 500, 502, 503, 504])

const now = Date.now()

const mockSessionStore = new Map<string, SessionDetail>([
  [
    "mock-clawpet-001",
    {
      id: "mock-clawpet-001",
      created: new Date(now - 1000 * 60 * 60 * 4).toISOString(),
      updated: new Date(now - 1000 * 60 * 10).toISOString(),
      summary: "Product planning and frontend implementation notes",
      messages: [
        {
          role: "user",
          content: "Help me plan the next ClawPet milestone.",
        },
        {
          role: "assistant",
          content:
            "Great idea. We can split this into three tracks: interaction design, pet-memory loop, and tool integration.",
        },
        {
          role: "user",
          content: "Can you draft API-first tasks that we can ship this week?",
        },
        {
          role: "assistant",
          content:
            "Yes. Ship list: 1) tools settings endpoint integration, 2) session history hydration fallback, 3) chat event parser for updates/media.",
        },
      ],
    },
  ],
  [
    "mock-clawpet-002",
    {
      id: "mock-clawpet-002",
      created: new Date(now - 1000 * 60 * 60 * 26).toISOString(),
      updated: new Date(now - 1000 * 60 * 60 * 2).toISOString(),
      summary: "Tool capability check",
      messages: [
        {
          role: "user",
          content: "Which tools are enabled right now?",
        },
        {
          role: "assistant",
          content:
            "Filesystem and web tools are enabled in this mock session. You can also toggle them on the Tools page.",
        },
      ],
    },
  ],
])

const mockSessionOrder = ["mock-clawpet-001", "mock-clawpet-002"]

let sessionsFallbackActive = false

function shouldFallbackToMock(error: unknown): boolean {
  if (!ENABLE_MOCK_FALLBACK) {
    return false
  }

  if (error instanceof ApiRequestError) {
    return FALLBACK_STATUS_CODES.has(error.status)
  }

  return error instanceof Error
}

function cloneSessionDetail(detail: SessionDetail): SessionDetail {
  return {
    ...detail,
    messages: detail.messages.map((message) => ({
      ...message,
      media: message.media ? [...message.media] : undefined,
    })),
  }
}

function buildSessionSummary(detail: SessionDetail): SessionSummary {
  const firstUser = detail.messages.find((message) => message.role === "user")
  const preview = (firstUser?.content || detail.summary || "(empty)").slice(0, 120)

  return {
    id: detail.id,
    title: preview,
    preview,
    message_count: detail.messages.length,
    created: detail.created,
    updated: detail.updated,
  }
}

function getMockSessions(offset: number, limit: number): SessionSummary[] {
  const summaries = mockSessionOrder
    .map((id) => mockSessionStore.get(id))
    .filter((detail): detail is SessionDetail => Boolean(detail))
    .map((detail) => buildSessionSummary(detail))

  return summaries.slice(offset, offset + limit)
}

function getMockSessionByID(id: string): SessionDetail {
  const detail = mockSessionStore.get(id)
  if (detail) {
    return cloneSessionDetail(detail)
  }

  return {
    id,
    messages: [],
    summary: "",
    created: new Date().toISOString(),
    updated: new Date().toISOString(),
  }
}

function toHttpError(res: Response, message: string): ApiRequestError {
  return new ApiRequestError(res.status, message)
}

export async function getSessions(
  offset: number = 0,
  limit: number = 20,
): Promise<SessionSummary[]> {
  const params = new URLSearchParams({
    offset: offset.toString(),
    limit: limit.toString(),
  })

  try {
    const res = await launcherFetch(`/api/sessions?${params.toString()}`)
    if (!res.ok) {
      throw toHttpError(res, `Failed to fetch sessions: ${res.status}`)
    }

    sessionsFallbackActive = false
    return res.json()
  } catch (error) {
    if (shouldFallbackToMock(error)) {
      sessionsFallbackActive = true
      return getMockSessions(offset, limit)
    }
    throw error
  }
}

export async function getSessionHistory(id: string): Promise<SessionDetail> {
  try {
    const res = await launcherFetch(`/api/sessions/${encodeURIComponent(id)}`)
    if (!res.ok) {
      throw toHttpError(res, `Failed to fetch session ${id}: ${res.status}`)
    }

    sessionsFallbackActive = false
    return res.json()
  } catch (error) {
    if (sessionsFallbackActive || shouldFallbackToMock(error)) {
      sessionsFallbackActive = true
      return getMockSessionByID(id)
    }
    throw error
  }
}

export async function deleteSession(id: string): Promise<void> {
  try {
    const res = await launcherFetch(`/api/sessions/${encodeURIComponent(id)}`, {
      method: "DELETE",
    })
    if (!res.ok) {
      throw toHttpError(res, `Failed to delete session ${id}: ${res.status}`)
    }

    sessionsFallbackActive = false
    return
  } catch (error) {
    if (sessionsFallbackActive || shouldFallbackToMock(error)) {
      sessionsFallbackActive = true
      mockSessionStore.delete(id)
      const index = mockSessionOrder.indexOf(id)
      if (index >= 0) {
        mockSessionOrder.splice(index, 1)
      }
      return
    }
    throw error
  }
}
