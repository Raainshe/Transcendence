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
}
