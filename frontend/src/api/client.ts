import type { ApiErrorBody } from '@/types/api'

export const API_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? ''
const API_PREFIX = '/api/v1'

/** Resolve avatar or upload paths from the API for use in img src. */
export function resolveAssetUrl(path: string | null | undefined): string | null {
  if (!path) return null
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  return `${API_BASE}${path.startsWith('/') ? path : `/${path}`}`
}

let bearerToken: string | null = null

export function setBearerToken(token: string | null): void {
  bearerToken = token
}

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export type ApiFetchOptions = RequestInit & {
  /** When false, omit Authorization even if a token is set. Default true. */
  auth?: boolean
  /** HTTP statuses that resolve as null instead of throwing (expected failures). */
  allowStatuses?: number[]
}

export async function apiFetch<T>(path: string, options: ApiFetchOptions = {}): Promise<T> {
  const { auth = true, allowStatuses = [], ...init } = options
  const headers = new Headers(init.headers)

  const isFormData = typeof FormData !== 'undefined' && init.body instanceof FormData
  const hasJsonBody =
    !isFormData &&
    init.body !== undefined &&
    init.body !== null &&
    typeof init.body === 'string' &&
    !headers.has('Content-Type')
  if (hasJsonBody) {
    headers.set('Content-Type', 'application/json')
  }

  if (auth && bearerToken) {
    headers.set('Authorization', `Bearer ${bearerToken}`)
  }

  const url = `${API_BASE}${API_PREFIX}${path}`
  const response = await fetch(url, { ...init, headers })

  if (response.status === 204) {
    return undefined as T
  }

  const contentType = response.headers.get('Content-Type') ?? ''
  const isJson = contentType.includes('application/json')
  const payload = isJson ? ((await response.json()) as T & ApiErrorBody) : null

  if (!response.ok) {
    if (allowStatuses.includes(response.status)) {
      return null as T
    }
    let message =
      payload && typeof payload === 'object' && 'error' in payload && payload.error
        ? String(payload.error)
        : response.statusText || 'request failed'
    if (response.status === 502) {
      message =
        'Cannot reach the API server. Is the backend running? (expected at port 8080 for local dev)'
    }
    throw new ApiError(response.status, message)
  }

  return payload as T
}
