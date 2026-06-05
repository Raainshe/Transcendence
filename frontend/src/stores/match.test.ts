import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import { useMatchStore } from '@/stores/match'
import type { User } from '@/types/api'

vi.mock('@/composables/useMatchSocket', () => {
  const connected = { value: false }
  return {
    WS_TYPE_ERROR: 'error',
    WS_TYPE_PLAYER_STATE: 'player.state',
    useMatchSocket: () => ({
      connected,
      connect: vi.fn(),
      disconnect: vi.fn(),
      retryPendingConnect: vi.fn(),
      sendPlayerState: vi.fn(),
    }),
  }
})

vi.mock('@/api/matches', () => ({
  getMatch: vi.fn(),
}))

const SELF_ID = '11111111-1111-1111-1111-111111111111'
const OPP_ID = '22222222-2222-2222-2222-222222222222'

function testUser(id: string, username: string): User {
  return {
    id,
    username,
    email: `${username}@example.com`,
    avatar_url: null,
    role: 'user',
    is_2fa_enabled: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    last_seen_at: null,
  }
}

describe('useMatchStore opponentList', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    const auth = useAuthStore()
    auth.$patch({
      token: 'test-token',
      user: testUser(SELF_ID, 'self'),
      status: 'authenticated',
    })
  })

  it('seeds opponent slots from roster before websocket state arrives', () => {
    const store = useMatchStore()
    store.players = [
      { user_id: SELF_ID, username: 'self', avatar_url: null },
      { user_id: OPP_ID, username: 'rival', avatar_url: '/avatars/rival.png' },
    ]

    const list = store.opponentList
    expect(list).toHaveLength(1)
    expect(list[0]?.user_id).toBe(OPP_ID)
    expect(list[0]?.username).toBe('rival')
    expect(list[0]?.score).toBe(0)
    expect(list[0]?.lines).toBe(0)
    expect(list[0]?.level).toBe(1)
    expect(list[0]?.alive).toBe(true)
    expect(list[0]?.board).toBe('')
  })

  it('merges live player.state over roster placeholder', () => {
    const store = useMatchStore()
    store.players = [
      { user_id: SELF_ID, username: 'self', avatar_url: null },
      { user_id: OPP_ID, username: 'rival', avatar_url: null },
    ]

    store.applyOpponentState({
      user_id: OPP_ID,
      score: 1200,
      lines: 8,
      level: 3,
      alive: true,
      board: 'dGVzdA==',
    })

    const list = store.opponentList
    expect(list).toHaveLength(1)
    expect(list[0]?.score).toBe(1200)
    expect(list[0]?.lines).toBe(8)
    expect(list[0]?.level).toBe(3)
    expect(list[0]?.board).toBe('dGVzdA==')
    expect(list[0]?.username).toBe('rival')
  })

  it('seeds roster from match start handoff before websocket updates', () => {
    const store = useMatchStore()
    store.seedFromMatchStart(
      'game-1',
      [
        { user_id: SELF_ID, username: 'self', avatar_url: null },
        { user_id: OPP_ID, username: 'rival', avatar_url: null },
      ],
      424242,
    )

    expect(store.players).toHaveLength(2)
    expect(store.opponentList).toHaveLength(1)
    expect(store.opponentList[0]?.username).toBe('rival')
    expect(store.getSharedSeed()).toBe(424242)
  })

  it('ignores own player.state broadcasts', () => {
    const store = useMatchStore()
    store.players = [
      { user_id: SELF_ID, username: 'self', avatar_url: null },
      { user_id: OPP_ID, username: 'rival', avatar_url: null },
    ]

    store.applyOpponentState({
      user_id: SELF_ID,
      score: 999,
      lines: 99,
      level: 9,
      alive: true,
      board: 'c2VsZg==',
    })

    expect(store.opponents.size).toBe(0)
    expect(store.opponentList).toHaveLength(1)
    expect(store.opponentList[0]?.board).toBe('')
  })
})
