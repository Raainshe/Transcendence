import { onUnmounted, ref } from 'vue'

import { useAuthStore } from '@/stores/auth'

export type WsEnvelope = {
  type: string
  payload?: unknown
}

function wsBaseUrl(): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/v1/ws`
}

export function useChatSocket() {
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

  function connect(handler: (env: WsEnvelope) => void): void {
    disconnect()
    onEnvelope = handler
    const token = auth.token
    if (!token) return
    socket = new WebSocket(`${wsBaseUrl()}?token=${encodeURIComponent(token)}`)
    socket.onopen = () => {
      connected.value = true
    }
    socket.onmessage = (event) => {
      try {
        onEnvelope?.(JSON.parse(String(event.data)) as WsEnvelope)
      } catch {
      }
    }
    socket.onclose = () => {
      connected.value = false
    }
    socket.onerror = () => {
      connected.value = false
    }
  }

  function send(to: string, body: string): boolean {
    if (!socket || socket.readyState !== WebSocket.OPEN) return false
    socket.send(JSON.stringify({ type: 'chat.send', payload: { to, body } }))
    return true
  }

  onUnmounted(() => {
    disconnect()
  })

  return { connected, connect, disconnect, send }
}
