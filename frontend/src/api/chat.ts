import { apiFetch } from '@/api/client'
import type {
  ConversationResponse,
  UnreadResponse
} from '@/types/api'

export function getConversation(friendID: string, limit = 50): Promise<ConversationResponse> {
  return apiFetch<ConversationResponse>(`/chat/messages/${friendID}?limit=${limit}`, { method: 'GET' })
}

export function markConversationRead(friendID: string): Promise<void> {
  return apiFetch<void>(`/chat/messages/${friendID}/read`, { method: 'POST' })
}

export function getUnread(): Promise<UnreadResponse> {
  return apiFetch<UnreadResponse>('/chat/unread', { method: 'GET' })
}
