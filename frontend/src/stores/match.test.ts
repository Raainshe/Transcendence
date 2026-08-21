import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import { useGameSessionStore } from '@/stores/gameSession'
import { useMatchStore } from '@/stores/match'
import type { User } from '@/types/api'

const sendPlayerEliminated = vi.fn()
const sendPlayerState = vi.fn()

vi.mock('@/composables/useMatchSocket', () => {
  const connected = { value: true }
  const reconnecting = { value: false }
  return {
    WS_TYPE_ERROR: 'error',
    WS_TYPE_PLAYER_STATE: 'player.state',
    WS_TYPE_PLAYER_DISCONNECTED: 'player.disconnected',
    WS_TYPE_PLAYER_RECONNECTED: 'player.reconnected',
    WS_TYPE_PLAYER_ELIMINATED: 'player.eliminated',
    WS_TYPE_MATCH_ENDED: 'match.ended',
    useMatchSocket: () => ({
      connected,
      reconnecting,
      connect: vi.fn(),
      disconnect: vi.fn(),
      retryPendingConnect: vi.fn(),
      sendPlayerState,
      sendPlayerEliminated,
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
    sendPlayerEliminated.mockClear()
    sendPlayerState.mockClear()
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

  it('marks opponent eliminated from player.eliminated broadcast', () => {
    const store = useMatchStore()
    store.players = [
      { user_id: SELF_ID, username: 'self', avatar_url: null },
      { user_id: OPP_ID, username: 'rival', avatar_url: null },
    ]
    store.applyOpponentState({
      user_id: OPP_ID,
      score: 300,
      lines: 2,
      level: 2,
      alive: true,
      board: 'dGVzdA==',
    })

    store.applyOpponentState({
      user_id: OPP_ID,
      score: 300,
      lines: 2,
      level: 2,
      alive: false,
      board: 'dGVzdA==',
    })

    expect(store.opponentList[0]?.alive).toBe(false)
  })

  it('stores match.ended results and stops the session', () => {
    const store = useMatchStore()
    store.gameId = 'game-1'
    const session = useGameSessionStore()
    session.multiplayerGameId = 'game-1'
    session.active = true

    store.applyMatchEnded({
      winner_user_id: OPP_ID,
      players: [
        {
          user_id: SELF_ID,
          username: 'self',
          score: 100,
          lines: 1,
          level: 1,
          placement: 2,
          is_winner: false,
        },
        {
          user_id: OPP_ID,
          username: 'rival',
          score: 500,
          lines: 5,
          level: 2,
          placement: 1,
          is_winner: true,
        },
      ],
    })

    expect(store.matchEnded).toBe(true)
    expect(store.matchResults?.winner_user_id).toBe(OPP_ID)
    expect(session.matchEndKind).toBe('lost')
    expect(session.active).toBe(false)
  })

  it('marks opponent disconnected and reconnected from broadcasts', () => {
    const store = useMatchStore()
    store.players = [
      { user_id: SELF_ID, username: 'self', avatar_url: null },
      { user_id: OPP_ID, username: 'rival', avatar_url: null },
    ]

    store.applyOpponentDisconnected({ user_id: OPP_ID })
    expect(store.opponentList[0]?.connected).toBe(false)

    store.applyOpponentReconnected({ user_id: OPP_ID })
    expect(store.opponentList[0]?.connected).toBe(true)
  })

  it('bootstrap with finished match applies stored results', async () => {
    const { getMatch } = await import('@/api/matches')
    vi.mocked(getMatch).mockResolvedValue({
      game_id: 'game-finished',
      status: 'finished',
      mode: 'multiplayer',
      lobby_name: 'Friday Night Tetris',
      shared_seed: 123,
      players: [
        { user_id: SELF_ID, username: 'self', avatar_url: null },
        { user_id: OPP_ID, username: 'rival', avatar_url: null },
      ],
      results: {
        winner_user_id: OPP_ID,
        players: [
          {
            user_id: SELF_ID,
            username: 'self',
            score: 100,
            lines: 1,
            level: 1,
            placement: 2,
            is_winner: false,
          },
          {
            user_id: OPP_ID,
            username: 'rival',
            score: 500,
            lines: 5,
            level: 2,
            placement: 1,
            is_winner: true,
          },
        ],
      },
    })

    const store = useMatchStore()
    const ok = await store.bootstrap('game-finished')

    expect(ok).toBe(false)
    expect(store.matchEnded).toBe(true)
    expect(store.matchResults?.winner_user_id).toBe(OPP_ID)
  })

  it('notifyLocalElimination sends only once', () => {
    const store = useMatchStore()
    store.gameId = 'game-1'

    store.notifyLocalElimination('topOut', { score: 50, lines: 1, level: 1 })
    store.notifyLocalElimination('topOut', { score: 50, lines: 1, level: 1 })

    expect(sendPlayerEliminated).toHaveBeenCalledTimes(1)
    expect(sendPlayerEliminated).toHaveBeenCalledWith('game-1', {
      reason: 'topOut',
      score: 50,
      lines: 1,
      level: 1,
    })
  })
})
