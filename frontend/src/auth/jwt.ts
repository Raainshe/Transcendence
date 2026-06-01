/** Refresh this long before JWT `exp` (backend TTL is 15 minutes). */
export const REFRESH_LEAD_MS = 60_000

type JwtPayload = {
  exp?: number
}

function decodeBase64Url(segment: string): string {
  const base64 = segment.replace(/-/g, '+').replace(/_/g, '/')
  const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=')
  return atob(padded)
}

/** Returns JWT expiry as Unix ms, or null if the token cannot be parsed. */
export function getTokenExpiryMs(token: string): number | null {
  const parts = token.split('.')
  const payloadSegment = parts[1]
  if (parts.length !== 3 || !payloadSegment) return null
  try {
    const payload = JSON.parse(decodeBase64Url(payloadSegment)) as JwtPayload
    if (typeof payload.exp !== 'number' || !Number.isFinite(payload.exp)) return null
    return payload.exp * 1000
  } catch {
    return null
  }
}

/** Ms until refresh should run (exp minus lead time). Negative means refresh now. */
export function msUntilRefresh(token: string, leadMs = REFRESH_LEAD_MS): number | null {
  const expMs = getTokenExpiryMs(token)
  if (expMs === null) return null
  return expMs - leadMs - Date.now()
}

export function shouldRefreshNow(token: string, leadMs = REFRESH_LEAD_MS): boolean {
  const remaining = msUntilRefresh(token, leadMs)
  return remaining === null || remaining <= 0
}
