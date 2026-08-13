import { apiFetch } from '@/api/client'
import type {
  CreateLobbyPayload,
  CurrentLobbyResponse,
  JoinLobbyByCodePayload,
  LobbyResponse,
  SetReadyPayload,
  StartLobbyResult,
} from '@/types/api'

export function createLobby(payload: CreateLobbyPayload): Promise<LobbyResponse> {
  return apiFetch<LobbyResponse>('/lobbies', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function joinLobbyByCode(payload: JoinLobbyByCodePayload): Promise<LobbyResponse | null> {
  return apiFetch<LobbyResponse | null>('/lobbies/join', {
    method: 'POST',
    body: JSON.stringify(payload),
    allowStatuses: [404],
  })
}

export function getLobby(id: string): Promise<LobbyResponse | null> {
  return apiFetch<LobbyResponse | null>(`/lobbies/${id}`, {
    method: 'GET',
    allowStatuses: [404],
  })
}

export function getCurrentLobby(): Promise<CurrentLobbyResponse> {
  return apiFetch<CurrentLobbyResponse>('/lobbies/current', { method: 'GET' })
}

export function setReady(id: string, payload: SetReadyPayload): Promise<LobbyResponse> {
  return apiFetch<LobbyResponse>(`/lobbies/${id}/ready`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function startLobby(id: string): Promise<StartLobbyResult> {
  return apiFetch<StartLobbyResult>(`/lobbies/${id}/start`, { method: 'POST' })
}

export function leaveLobby(id: string): Promise<{ status: string }> {
  return apiFetch<{ status: string }>(`/lobbies/${id}/leave`, { method: 'DELETE' })
}
