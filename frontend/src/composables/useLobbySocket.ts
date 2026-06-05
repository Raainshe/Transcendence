import { onUnmounted, ref } from 'vue'

import { useAuthStore } from '@/stores/auth'

export const WS_TYPE_LOBBY_UPDATED = 'lobby.updated'
export const WS_TYPE_LOBBY_CLOSED = 'lobby.closed'
export const WS_TYPE_MATCH_START = 'match.start'
export const WS_TYPE_ERROR = 'error'

export type WsEnvelope = {
  type: string
  payload?: unknown
}

export type LobbyClosedPayload = {
  reason: 'host_left' | 'started'
}

export type MatchStartPayload = {
  game_id: string
  shared_seed: number
  players: Array<{
    user_id: string
    username: string
    avatar_url: string | null
  }>
}

function wsBaseUrl(): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/v1/ws`
}

export function useLobbySocket() {
  const auth = useAuthStore()
  const connected = ref(false)
  let socket: WebSocket | null = null
  let onEnvelope: ((env: WsEnvelope) => void) | null = null

  function disconnect(): void {
    if (socket) {
      socket.onopen = null
      socket.onmessage = null
      socket.onclose = null
      socket.onerror = null
      socket.close()
      socket = null
    }
    connected.value = false
    onEnvelope = null
  }

  function connect(id: string, handler: (env: WsEnvelope) => void): void {
    disconnect()
    onEnvelope = handler

    const token = auth.token
    if (!token) return

    const url = `${wsBaseUrl()}?token=${encodeURIComponent(token)}`
    socket = new WebSocket(url)

    socket.onopen = () => {
      connected.value = true
      socket?.send(
        JSON.stringify({
          type: 'subscribe',
          room: `lobby:${id}`,
        }),
      )
    }

    socket.onmessage = (event) => {
      try {
        const env = JSON.parse(String(event.data)) as WsEnvelope
        onEnvelope?.(env)
      } catch {
        /* ignore malformed messages */
      }
    }

    socket.onclose = () => {
      connected.value = false
    }

    socket.onerror = () => {
      connected.value = false
    }
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    connect,
    disconnect,
  }
}
