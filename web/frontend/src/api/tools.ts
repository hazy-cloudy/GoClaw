import { launcherFetch } from "@/api/http"

export interface ToolSupportItem {
  name: string
  description: string
  category: string
  config_key: string
  status: "enabled" | "disabled" | "blocked"
  reason_code?: string
}

export type ToolDataSource = "live" | "mock"

export interface ToolsResponse {
  tools: ToolSupportItem[]
  source: ToolDataSource
}

export interface ToolActionResponse {
  status: string
  source: ToolDataSource
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

const MOCK_TOOL_CATALOG: ToolSupportItem[] = [
  {
    name: "read_file",
    description: "Read file content from the workspace or explicitly allowed paths.",
    category: "filesystem",
    config_key: "read_file",
    status: "enabled",
  },
  {
    name: "write_file",
    description: "Create or overwrite files within the writable workspace scope.",
    category: "filesystem",
    config_key: "write_file",
    status: "enabled",
  },
  {
    name: "list_dir",
    description: "Inspect directories and enumerate files available to the agent.",
    category: "filesystem",
    config_key: "list_dir",
    status: "enabled",
  },
  {
    name: "edit_file",
    description: "Apply targeted edits to existing files without rewriting everything.",
    category: "filesystem",
    config_key: "edit_file",
    status: "enabled",
  },
  {
    name: "append_file",
    description: "Append content to the end of an existing file.",
    category: "filesystem",
    config_key: "append_file",
    status: "disabled",
  },
  {
    name: "exec",
    description: "Run shell commands inside the configured workspace sandbox.",
    category: "filesystem",
    config_key: "exec",
    status: "enabled",
  },
  {
    name: "cron",
    description: "Schedule one-time or recurring reminders, jobs, and shell commands.",
    category: "automation",
    config_key: "cron",
    status: "enabled",
  },
  {
    name: "web_search",
    description: "Search the web using the configured providers.",
    category: "web",
    config_key: "web",
    status: "enabled",
  },
  {
    name: "web_fetch",
    description: "Fetch and summarize the contents of a webpage.",
    category: "web",
    config_key: "web_fetch",
    status: "enabled",
  },
  {
    name: "message",
    description: "Send a follow-up message back to the active user or chat.",
    category: "communication",
    config_key: "message",
    status: "enabled",
  },
  {
    name: "send_file",
    description: "Send an outbound file or media attachment to the active chat.",
    category: "communication",
    config_key: "send_file",
    status: "enabled",
  },
  {
    name: "find_skills",
    description: "Search external skill registries for installable skills.",
    category: "skills",
    config_key: "find_skills",
    status: "disabled",
  },
  {
    name: "install_skill",
    description: "Install a skill into the current workspace from a registry.",
    category: "skills",
    config_key: "install_skill",
    status: "disabled",
  },
  {
    name: "spawn",
    description: "Launch a background subagent for long-running or delegated work.",
    category: "agents",
    config_key: "spawn",
    status: "enabled",
  },
  {
    name: "spawn_status",
    description: "Query the status of spawned subagents.",
    category: "agents",
    config_key: "spawn_status",
    status: "enabled",
  },
  {
    name: "i2c",
    description: "Interact with I2C hardware devices exposed on the host.",
    category: "hardware",
    config_key: "i2c",
    status: "blocked",
    reason_code: "requires_linux",
  },
  {
    name: "spi",
    description: "Interact with SPI hardware devices exposed on the host.",
    category: "hardware",
    config_key: "spi",
    status: "blocked",
    reason_code: "requires_linux",
  },
]

const mockToolStates = new Map(
  MOCK_TOOL_CATALOG.map((tool) => [tool.name, tool.status]),
)

let toolsFallbackActive = false

function shouldFallbackToMock(error: unknown): boolean {
  if (!ENABLE_MOCK_FALLBACK) {
    return false
  }

  if (error instanceof ApiRequestError) {
    return FALLBACK_STATUS_CODES.has(error.status)
  }

  return error instanceof Error
}

function buildMockTools(): ToolSupportItem[] {
  return MOCK_TOOL_CATALOG.map((tool) => {
    const status = mockToolStates.get(tool.name) ?? tool.status
    return {
      ...tool,
      status,
      reason_code:
        status === "blocked" && (tool.name === "i2c" || tool.name === "spi")
          ? "requires_linux"
          : undefined,
    }
  })
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await launcherFetch(path, options)
  if (!res.ok) {
    let message = `API error: ${res.status} ${res.statusText}`
    try {
      const body = (await res.json()) as {
        error?: string
        errors?: string[]
      }
      if (Array.isArray(body.errors) && body.errors.length > 0) {
        message = body.errors.join("; ")
      } else if (typeof body.error === "string" && body.error.trim() !== "") {
        message = body.error
      }
    } catch {
      // ignore invalid body
    }
    throw new ApiRequestError(res.status, message)
  }
  return res.json() as Promise<T>
}

export async function getTools(): Promise<ToolsResponse> {
  try {
    const response = await request<{ tools: ToolSupportItem[] }>("/api/tools")
    toolsFallbackActive = false
    return {
      tools: response.tools,
      source: "live",
    }
  } catch (error) {
    if (shouldFallbackToMock(error)) {
      toolsFallbackActive = true
      return {
        tools: buildMockTools(),
        source: "mock",
      }
    }
    throw error
  }
}

export async function setToolEnabled(
  name: string,
  enabled: boolean,
): Promise<ToolActionResponse> {
  try {
    await request<{ status: string }>(`/api/tools/${encodeURIComponent(name)}/state`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ enabled }),
    })
    toolsFallbackActive = false
    return { status: "ok", source: "live" }
  } catch (error) {
    const canUseMock = toolsFallbackActive || shouldFallbackToMock(error)
    if (!canUseMock) {
      throw error
    }

    if (mockToolStates.has(name)) {
      let nextStatus: ToolSupportItem["status"] = enabled ? "enabled" : "disabled"
      if (enabled && (name === "i2c" || name === "spi")) {
        nextStatus = "blocked"
      }
      mockToolStates.set(name, nextStatus)
    }

    toolsFallbackActive = true
    return { status: "ok", source: "mock" }
  }
}
