import { apiFetch } from '@/api/client'
import type {
  BlockedUsersResponse,
  FriendRequestsResponse,
  FriendsResponse,
} from '@/types/api'

export function getFriends(): Promise<FriendsResponse> {
  return apiFetch<FriendsResponse>('/users/me/friends', { method: 'GET' })
}

export function getFriendRequests(): Promise<FriendRequestsResponse> {
  return apiFetch<FriendRequestsResponse>('/users/me/friends/requests', { method: 'GET' })
}

export function getBlockedUsers(): Promise<BlockedUsersResponse> {
  return apiFetch<BlockedUsersResponse>('/users/me/blocks', { method: 'GET' })
}

export function sendFriendRequest(userId: string): Promise<void> {
  return apiFetch<void>(`/users/me/friends/${userId}`, { method: 'POST' })
}

export function acceptFriendRequest(userId: string): Promise<void> {
  return apiFetch<void>(`/users/me/friends/${userId}`, { method: 'PATCH' })
}

export function removeFriend(userId: string): Promise<void> {
  return apiFetch<void>(`/users/me/friends/${userId}`, { method: 'DELETE' })
}

export function blockUser(userId: string): Promise<void> {
  return apiFetch<void>(`/users/me/block/${userId}`, { method: 'POST' })
}

export function unblockUser(userId: string): Promise<void> {
  return apiFetch<void>(`/users/me/block/${userId}`, { method: 'DELETE' })
}
