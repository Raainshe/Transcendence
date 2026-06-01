import { msUntilRefresh } from '@/auth/jwt'

let refreshTimer: ReturnType<typeof setTimeout> | null = null

export function stopTokenRefreshTimer(): void {
  if (refreshTimer !== null) {
    clearTimeout(refreshTimer)
    refreshTimer = null
  }
}

export function scheduleTokenRefresh(
  token: string,
  onRefresh: () => void | Promise<void>,
): void {
  stopTokenRefreshTimer()

  const delay = msUntilRefresh(token)
  if (delay === null) return

  refreshTimer = setTimeout(
    () => {
      refreshTimer = null
      void onRefresh()
    },
    Math.max(0, delay),
  )
}
