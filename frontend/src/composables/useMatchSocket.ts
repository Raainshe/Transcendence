import { ref } from 'vue'

import { useAuthStore } from '@/stores/auth'
import type { PlayerEliminatedUpload, PlayerStateUpload } from '@/types/api'

export const WS_TYPE_PLAYER_STATE = 'player.state'
export const WS_TYPE_PLAYER_ELIMINATED = 'player.eliminated'
export const WS_TYPE_MATCH_ENDED = 'match.ended'
export const WS_TYPE_ERROR = 'error'

export type WsEnvelope = {
  type: string
  payload?: unknown
}

function wsBaseUrl(): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/v1/ws`
}

function matchRoom(gameId: string): string {
  return `match:${gameId}`
}

export function useMatchSocket() {
  const auth = useAuthStore()
  const connected = ref(false)
  let socket: WebSocket | null = null
  let onEnvelope: ((env: WsEnvelope) => void) | null = null
  let pendingGameId: string | null = null

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
    pendingGameId = null
  }

  function openSocket(gameId: string): void {
    const token = auth.token
    if (!token) {
      pendingGameId = gameId
      return
    }

    pendingGameId = null
    const url = `${wsBaseUrl()}?token=${encodeURIComponent(token)}`
    socket = new WebSocket(url)

    socket.onopen = () => {
      connected.value = true
      socket?.send(
        JSON.stringify({
          type: 'subscribe',
          room: matchRoom(gameId),
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

  function connect(gameId: string, handler: (env: WsEnvelope) => void): void {
    disconnect()
    onEnvelope = handler
    openSocket(gameId)
  }

  function retryPendingConnect(): void {
    if (!pendingGameId || !onEnvelope || !auth.token) return
    const gameId = pendingGameId
    if (socket) return
    openSocket(gameId)
  }

  function sendPlayerState(gameId: string, payload: PlayerStateUpload): void {
    if (!socket || socket.readyState !== WebSocket.OPEN) return
    socket.send(
      JSON.stringify({
        type: WS_TYPE_PLAYER_STATE,
        room: matchRoom(gameId),
        payload,
      }),
    )
  }

  function sendPlayerEliminated(gameId: string, payload: PlayerEliminatedUpload): void {
    if (!socket || socket.readyState !== WebSocket.OPEN) return
    socket.send(
      JSON.stringify({
        type: WS_TYPE_PLAYER_ELIMINATED,
        room: matchRoom(gameId),
        payload,
      }),
    )
  }

  return {
    connected,
    connect,
    disconnect,
    retryPendingConnect,
    sendPlayerState,
    sendPlayerEliminated,
  }
}
