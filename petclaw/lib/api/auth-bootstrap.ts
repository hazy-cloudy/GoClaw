import { API_ENDPOINTS, getApiBaseUrl, resolveLauncherToken } from './config'

let inflightBootstrap: Promise<boolean> | null = null

interface AuthStatusResponse {
  authenticated?: boolean
}

async function fetchAuthStatus(): Promise<boolean> {
  const res = await fetch(`${getApiBaseUrl()}${API_ENDPOINTS.AUTH.STATUS}`, {
    credentials: 'include',
  })

  if (!res.ok) {
    return false
  }

  const data = (await res.json().catch(() => ({}))) as AuthStatusResponse
  return Boolean(data.authenticated)
}

async function loginWithLauncherToken(token: string): Promise<boolean> {
  const res = await fetch(`${getApiBaseUrl()}${API_ENDPOINTS.AUTH.LOGIN}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({ token }),
  })
  return res.ok
}

export async function ensureLauncherAuthSession(): Promise<boolean> {
  if (typeof window === 'undefined') {
    return false
  }

  if (inflightBootstrap) {
    return inflightBootstrap
  }

  inflightBootstrap = (async () => {
    if (await fetchAuthStatus()) {
      return true
    }

    const token = resolveLauncherToken()
    if (!token) {
      return false
    }

    if (!(await loginWithLauncherToken(token))) {
      return false
    }

    return fetchAuthStatus()
  })()

  try {
    return await inflightBootstrap
  } finally {
    inflightBootstrap = null
  }
}

export async function fetchWithAuthRetry(input: string, init: RequestInit = {}, allowRetry = true): Promise<Response> {
  const res = await fetch(input, init)
  if (!allowRetry || res.status !== 401) {
    return res
  }

  const recovered = await ensureLauncherAuthSession()
  if (!recovered) {
    return res
  }

  return fetch(input, init)
}
