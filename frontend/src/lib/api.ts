'use client'

import { refreshToken } from '@/app/actions/auth'

const BASE = '/api/v1'

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

let refreshPromise: Promise<boolean> | null = null

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const isFormData = init.body instanceof FormData
  const { headers, ...rest } = init

  const buildInit = (): RequestInit => ({
    credentials: 'include',
    ...rest,
    headers: isFormData
      ? (headers as HeadersInit | undefined)
      : { 'Content-Type': 'application/json', ...(headers as Record<string, string>) },
  })

  let res = await fetch(`${BASE}${path}`, buildInit())

  if (res.status === 401) {
    if (!refreshPromise) {
      refreshPromise = refreshToken().finally(() => { refreshPromise = null })
    }
    const refreshed = await refreshPromise
    if (!refreshed) {
      window.location.href = '/login'
      throw new ApiError('Session expirée', 401)
    }
    res = await fetch(`${BASE}${path}`, buildInit())
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(body.error ?? String(res.status), res.status)
  }

  return res.json() as Promise<T>
}

export const apiClient = {
  get: <T>(path: string, init?: RequestInit) =>
    request<T>(path, { ...init, method: 'GET' }),

  post: <T>(path: string, body?: unknown, init?: RequestInit) =>
    request<T>(path, {
      ...init,
      method: 'POST',
      body: body instanceof FormData ? body : JSON.stringify(body),
    }),

  put: <T>(path: string, body?: unknown, init?: RequestInit) =>
    request<T>(path, { ...init, method: 'PUT', body: JSON.stringify(body) }),

  delete: <T = void>(path: string, init?: RequestInit) =>
    request<T>(path, { ...init, method: 'DELETE' }),

  head: (path: string, init?: RequestInit) =>
    fetch(`${BASE}${path}`, { credentials: 'include', ...init, method: 'HEAD' }),
}
