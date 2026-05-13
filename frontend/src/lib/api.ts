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

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const isFormData = init.body instanceof FormData
  const { headers, ...rest } = init

  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include',
    ...rest,
    headers: isFormData
      ? (headers as HeadersInit | undefined)
      : { 'Content-Type': 'application/json', ...(headers as Record<string, string>) },
  })

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
