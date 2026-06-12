import type { ApiResponse, ApiErrorResponse } from '@/types/api'

function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)kora_csrf=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : ''
}

export class KoraApiError extends Error {
  type: string
  field?: string
  status: number

  constructor(message: string, type: string, status: number, field?: string) {
    super(message)
    this.name = 'KoraApiError'
    this.type = type
    this.status = status
    this.field = field
  }
}

async function apiRequest<T>(
  method: string,
  path: string,
  body?: unknown,
  params?: Record<string, string | number | undefined>,
): Promise<T> {
  const url = new URL(path, window.location.origin)
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined) {
        url.searchParams.set(key, String(value))
      }
    }
  }

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  // CSRF token for state-changing requests.
  if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
    const csrf = getCsrfToken()
    if (csrf) {
      headers['X-Kora-CSRF-Token'] = csrf
    }
  }

  const response = await fetch(url.toString(), {
    method,
    headers,
    credentials: 'same-origin',
    body: body ? JSON.stringify(body) : undefined,
  })

  if (!response.ok) {
    let errorData: ApiErrorResponse
    try {
      errorData = await response.json()
    } catch {
      throw new KoraApiError(
        `Request failed with status ${response.status}`,
        'network_error',
        response.status,
      )
    }

    // Plain string error: {"error": "some message"}
    if (typeof errorData.error === 'string') {
      throw new KoraApiError(errorData.error, 'error', response.status)
    }
    // Single error: {"error": {"message": "...", "field": "..."}}
    if (errorData.error && typeof errorData.error === 'object' && 'message' in errorData.error) {
      const err = errorData.error as any
      throw new KoraApiError(err.message, err.type || 'error', response.status, err.field)
    }
    // Multiple errors: {"error": {"errors": [...]}}
    if (errorData.error && typeof errorData.error === 'object' && 'errors' in errorData.error) {
      const err = (errorData.error as any).errors[0]
      throw new KoraApiError(err.message, err.type || 'error', response.status, err.field)
    }
    throw new KoraApiError('Unknown error', 'unknown', response.status)
  }

  // No content.
  if (response.status === 204) {
    return undefined as T
  }

  const json: ApiResponse<T> = await response.json()
  return json.data
}

// Convenience methods.
export const api = {
  get<T>(path: string, params?: Record<string, string | number | undefined>) {
    return apiRequest<T>('GET', path, undefined, params)
  },
  // Returns the full response envelope including meta.
  getEnvelope<T>(path: string, params?: Record<string, string | number | undefined>) {
    return apiRequestEnvelope<T>('GET', path, undefined, params)
  },
  post<T>(path: string, body?: unknown) {
    return apiRequest<T>('POST', path, body)
  },
  put<T>(path: string, body?: unknown) {
    return apiRequest<T>('PUT', path, body)
  },
  delete<T>(path: string) {
    return apiRequest<T>('DELETE', path)
  },
}

// Like apiRequest but returns the full envelope.
async function apiRequestEnvelope<T>(
  method: string,
  path: string,
  body?: unknown,
  params?: Record<string, string | number | undefined>,
): Promise<{ data: T; meta?: { total?: number; doctype?: string; config_version?: number } }> {
  const url = new URL(path, window.location.origin)
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined) {
        url.searchParams.set(key, String(value))
      }
    }
  }

  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
    const csrf = getCsrfToken()
    if (csrf) headers['X-Kora-CSRF-Token'] = csrf
  }

  const response = await fetch(url.toString(), {
    method,
    headers,
    credentials: 'same-origin',
    body: body ? JSON.stringify(body) : undefined,
  })

  if (!response.ok) {
    let errorData: ApiErrorResponse
    try { errorData = await response.json() } catch {
      throw new KoraApiError(`Request failed with status ${response.status}`, 'network_error', response.status)
    }
    const err = (errorData.error as any)
    if (err?.message) throw new KoraApiError(err.message, err.type || 'error', response.status, err.field)
    if (err?.errors?.[0]) throw new KoraApiError(err.errors[0].message, err.errors[0].type || 'error', response.status, err.errors[0].field)
    throw new KoraApiError('Unknown error', 'unknown', response.status)
  }

  return response.json()
}
