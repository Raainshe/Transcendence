import { ApiError, apiFetch } from '@/api/client'
import type {
  MeResponse,
  UpdateMePayload,
  User,
  UserResponse,
  UsersListResponse,
  UserStats,
} from '@/types/api'

export function getUser(userId: string): Promise<User> {
  return apiFetch<UserResponse>(`/users/${userId}`, {
    method: 'GET',
    auth: false,
  }).then((res) => res.user)
}

export type ListUsersParams = {
  limit?: number
  offset?: number
}

export function listUsers(params: ListUsersParams = {}): Promise<UsersListResponse> {
  const search = new URLSearchParams()
  if (params.limit !== undefined && params.limit > 0) search.set('limit', String(params.limit))
  if (params.offset !== undefined && params.offset > 0) search.set('offset', String(params.offset))
  const query = search.toString()
  return apiFetch<UsersListResponse>(`/users${query ? `?${query}` : ''}`, {
    method: 'GET',
    auth: false,
  })
}

export function findUserByUsername(username: string): Promise<User | null> {
  const params = new URLSearchParams({ username })
  return apiFetch<UsersListResponse>(`/users?${params.toString()}`, {
    method: 'GET',
    auth: false,
  })
    .then((res) => res.users[0] ?? null)
    .catch((error: unknown) => {
      if (error instanceof ApiError && error.status === 404) return null
      throw error
    })
}

export function updateMe(payload: UpdateMePayload): Promise<MeResponse> {
  return apiFetch<MeResponse>('/users/me', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export function uploadAvatar(file: File): Promise<MeResponse> {
  const form = new FormData()
  form.append('avatar', file)
  return apiFetch<MeResponse>('/users/me/avatar', {
    method: 'POST',
    body: form,
  })
}

export function deleteAvatar(): Promise<void> {
  return apiFetch<void>('/users/me/avatar', {
    method: 'DELETE',
  })
}

export function deleteMe(): Promise<void> {
  return apiFetch<void>('/users/me', {
    method: 'DELETE',
  })
}

export function getUserStats(userId: string): Promise<UserStats> {
  return apiFetch<UserStats>(`/users/${userId}/stats`, {
    method: 'GET',
    auth: false,
  })
}
