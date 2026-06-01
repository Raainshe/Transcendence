import { ApiError, apiFetch } from '@/api/client'
import type { MeResponse, UpdateMePayload, User, UsersListResponse, UserStats } from '@/types/api'

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
