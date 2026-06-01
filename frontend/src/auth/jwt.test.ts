import { describe, expect, it } from 'vitest'

import { getTokenExpiryMs, msUntilRefresh, REFRESH_LEAD_MS, shouldRefreshNow } from '@/auth/jwt'

function makeToken(payload: Record<string, unknown>): string {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
  const body = btoa(JSON.stringify(payload))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
  return `${header}.${body}.signature`
}

describe('jwt helpers', () => {
  it('parses exp from a JWT payload', () => {
    const exp = Math.floor(Date.now() / 1000) + 900
    const token = makeToken({ exp, sub: 'user' })
    expect(getTokenExpiryMs(token)).toBe(exp * 1000)
  })

  it('returns null for malformed tokens', () => {
    expect(getTokenExpiryMs('not-a-jwt')).toBeNull()
    expect(getTokenExpiryMs('a.b')).toBeNull()
  })

  it('shouldRefreshNow when inside lead window', () => {
    const exp = Math.floor((Date.now() + REFRESH_LEAD_MS - 1000) / 1000)
    const token = makeToken({ exp })
    expect(shouldRefreshNow(token)).toBe(true)
    expect(msUntilRefresh(token)).toBeLessThanOrEqual(0)
  })

  it('should not refresh when far from expiry', () => {
    const exp = Math.floor((Date.now() + 10 * 60_000) / 1000)
    const token = makeToken({ exp })
    expect(shouldRefreshNow(token)).toBe(false)
    expect(msUntilRefresh(token)).toBeGreaterThan(0)
  })
})
