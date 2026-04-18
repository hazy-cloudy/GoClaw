const DEFAULT_API_BASE = 'http://127.0.0.1:18800'

export const API_ENDPOINTS = {
  AUTH: {
    STATUS: '/api/auth/status',
  },
  GATEWAY: {
    STATUS: '/api/gateway/status',
  },
  PET: {
    TOKEN: '/api/pet/token',
    SETUP: '/api/pet/setup',
  },
  PICO: {
    TOKEN: '/api/pico/token',
    SETUP: '/api/pico/setup',
  },
} as const

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

export function getApiBaseUrl(): string {
  const fromElectron = window.electronAPI?.getBackendBaseUrl?.()
  if (fromElectron && fromElectron.trim()) {
    return trimTrailingSlash(fromElectron.trim())
  }

  return DEFAULT_API_BASE
}

export async function fetchWithAuthRetry(input: string, init: RequestInit = {}, allowRetry = true): Promise<Response> {
  const first = await fetch(input, init)
  if (!allowRetry || first.status !== 401) {
    return first
  }

  return fetch(input, init)
}
