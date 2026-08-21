import { apiFetch } from '@/api/client'
import type { ApiKeyCreateResponse, ApiKeysResponse } from '@/types/api'

export function getApiKeys(): Promise<ApiKeysResponse> {
  return apiFetch<ApiKeysResponse>('/users/me/api-keys', { method: 'GET' })
}

export function createApiKey(name: string): Promise<ApiKeyCreateResponse> {
  return apiFetch<ApiKeyCreateResponse>('/users/me/api-keys', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export function revokeApiKey(id: string): Promise<void> {
  return apiFetch<void>(`/users/me/api-keys/${id}`, { method: 'DELETE' })
}
