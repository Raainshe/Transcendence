import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'

import { getMatch } from '@/api/matches'
import {
  WS_TYPE_ERROR,
  WS_TYPE_MATCH_ENDED,
  WS_TYPE_PLAYER_ELIMINATED,
  WS_TYPE_PLAYER_STATE,
  type WsEnvelope,
  useMatchSocket,
} from '@/composables/useMatchSocket'
import { encodeLockedMatrixBase64 } from '@/game/engine/matrixSnapshot'
import {
  readMatchPlayers,
  stashMatchPlayers,
  stashMultiplayerSeed,
} from '@/stores/lobby'
import { useAuthStore } from '@/stores/auth'
import { useGameSessionStore } from '@/stores/gameSession'
import type {
  MatchEndedPayload,
  MatchPlayerView,
  PlayerEliminatedBroadcast,
  PlayerStateBroadcast,
} from '@/types/api'

const SEND_INTERVAL_MS = 150

export type OpponentView = PlayerStateBroadcast & {
  username: string
  avatar_url: string | null
}

function normalizeUserId(id: string | null | undefined): string | null {
  if (!id) return null
  return id.trim().toLowerCase()
}

function isSelfPlayer(userId: string, selfId: string | null | undefined): boolean {
  const normalizedSelf = normalizeUserId(selfId)
  if (!normalizedSelf) return false
  return normalizeUserId(userId) === normalizedSelf
}

function placeholderOpponent(player: MatchPlayerView): OpponentView {
  return {
    user_id: player.user_id,
    username: player.username,
    avatar_url: player.avatar_url,
    score: 0,
    lines: 0,
    level: 1,
    alive: true,
    board: '',
  }
}

export const useMatchStore = defineStore('match', () => {
  const auth = useAuthStore()
  const socket = useMatchSocket()

  const gameId = ref<string | null>(null)
  const players = ref<MatchPlayerView[]>([])
  const opponents = ref<Map<string, OpponentView>>(new Map())
  const matchError = ref<string | null>(null)
  const matchEnded = ref(false)
  const matchResults = ref<MatchEndedPayload | null>(null)
  const connected = computed(() => socket.connected.value)
  const bootstrapFailed = ref(false)

  let lastSentAt = 0
  let sharedSeed = 0
  let stopTokenWatch: (() => void) | null = null
  let localEliminationSent = false

  const opponentList = computed(() => {
    const selfId = auth.user?.id
    const roster = players.value.filter((p) => !isSelfPlayer(p.user_id, selfId))

    if (roster.length > 0) {
      return roster.map((player) => {
        const live =
          opponents.value.get(player.user_id) ??
          [...opponents.value.values()].find(
            (o) => normalizeUserId(o.user_id) === normalizeUserId(player.user_id),
          )
        return live ?? placeholderOpponent(player)
      })
    }

    return [...opponents.value.values()].filter((o) => !isSelfPlayer(o.user_id, selfId))
  })

  function playerMeta(userId: string): { username: string; avatar_url: string | null } {
    const known = players.value.find(
      (p) => normalizeUserId(p.user_id) === normalizeUserId(userId),
    )
    if (known) {
      return { username: known.username, avatar_url: known.avatar_url }
    }
    return { username: userId.slice(0, 8), avatar_url: null }
  }

  function clearTokenWatch(): void {
    stopTokenWatch?.()
    stopTokenWatch = null
  }

  function reset(): void {
    gameId.value = null
    players.value = []
    opponents.value = new Map()
    matchError.value = null
    matchEnded.value = false
    matchResults.value = null
    bootstrapFailed.value = false
    lastSentAt = 0
    sharedSeed = 0
    localEliminationSent = false
    clearTokenWatch()
    socket.disconnect()
  }

  function applyStashedPlayers(id: string): void {
    const stashed = readMatchPlayers(id)
    if (stashed.length > 0) {
      players.value = stashed
    }
  }

  function ensurePlayersRoster(id: string): void {
    if (players.value.length === 0) {
      applyStashedPlayers(id)
    }
  }

  function seedFromMatchStart(
    id: string,
    roster: MatchPlayerView[],
    seed?: number,
  ): void {
    gameId.value = id
    if (roster.length > 0) {
      players.value = roster
      stashMatchPlayers(id, roster)
    }
    if (seed != null && seed > 0) {
      sharedSeed = seed
      stashMultiplayerSeed(id, seed)
    }
  }

  async function bootstrap(id: string): Promise<boolean> {
    const preservedPlayers = gameId.value === id && players.value.length > 0 ? [...players.value] : []
    socket.disconnect()
    opponents.value = new Map()
    matchError.value = null
    bootstrapFailed.value = false
    lastSentAt = 0
    if (gameId.value !== id) {
      sharedSeed = 0
      players.value = []
    }
    gameId.value = id

    try {
      const detail = await getMatch(id)
      if (!detail) {
        bootstrapFailed.value = true
        applyStashedPlayers(id)
        if (players.value.length === 0 && preservedPlayers.length > 0) {
          players.value = preservedPlayers
        }
        return false
      }

      players.value = detail.players ?? []
      if (players.value.length > 0) {
        stashMatchPlayers(id, players.value)
      }
      sharedSeed = detail.shared_seed
      if (sharedSeed > 0) {
        stashMultiplayerSeed(id, sharedSeed)
      }
      return true
    } catch {
      bootstrapFailed.value = true
      applyStashedPlayers(id)
      if (players.value.length === 0 && preservedPlayers.length > 0) {
        players.value = preservedPlayers
      }
      return false
    }
  }

  function getSharedSeed(): number {
    return sharedSeed
  }

  function applyOpponentState(payload: PlayerStateBroadcast): void {
    if (isSelfPlayer(payload.user_id, auth.user?.id)) return

    const meta = playerMeta(payload.user_id)
    const next = new Map(opponents.value)
    next.set(payload.user_id, {
      ...payload,
      username: meta.username,
      avatar_url: meta.avatar_url,
    })
    opponents.value = next
  }

  function applyOpponentEliminated(payload: PlayerEliminatedBroadcast): void {
    if (isSelfPlayer(payload.user_id, auth.user?.id)) return

    const meta = playerMeta(payload.user_id)
    const existing =
      opponents.value.get(payload.user_id) ??
      [...opponents.value.values()].find(
        (o) => normalizeUserId(o.user_id) === normalizeUserId(payload.user_id),
      )

    const next = new Map(opponents.value)
    next.set(payload.user_id, {
      user_id: payload.user_id,
      username: meta.username,
      avatar_url: meta.avatar_url,
      score: existing?.score ?? 0,
      lines: existing?.lines ?? 0,
      level: existing?.level ?? 1,
      alive: false,
      board: existing?.board ?? '',
    })
    opponents.value = next
  }

  function applyMatchEnded(payload: MatchEndedPayload): void {
    matchEnded.value = true
    matchResults.value = payload
    useGameSessionStore().applyServerMatchEnd(payload)
    socket.disconnect()
  }

  function handleEnvelope(env: WsEnvelope): void {
    if (env.type === WS_TYPE_ERROR) {
      const payload = env.payload as { error?: string } | undefined
      matchError.value = payload?.error ?? 'websocket error'
      return
    }
    if (env.type === WS_TYPE_PLAYER_ELIMINATED && env.payload) {
      applyOpponentEliminated(env.payload as PlayerEliminatedBroadcast)
      return
    }
    if (env.type === WS_TYPE_MATCH_ENDED && env.payload) {
      applyMatchEnded(env.payload as MatchEndedPayload)
      return
    }
    if (env.type === WS_TYPE_PLAYER_STATE && env.payload) {
      applyOpponentState(env.payload as PlayerStateBroadcast)
    }
  }

  function notifyLocalElimination(
    reason: string,
    stats: { score: number; lines: number; level: number },
  ): void {
    const id = gameId.value
    if (!id || matchEnded.value || localEliminationSent) return
    localEliminationSent = true

    socket.sendPlayerEliminated(id, {
      reason,
      score: stats.score,
      lines: stats.lines,
      level: stats.level,
    })

    const session = useGameSessionStore()
    const board = session.engine
      ? encodeLockedMatrixBase64(session.engine.matrixRef)
      : ''
    socket.sendPlayerState(id, {
      score: stats.score,
      lines: stats.lines,
      level: stats.level,
      alive: false,
      board,
    })
  }

  function connect(id: string): void {
    gameId.value = id
    matchError.value = null
    clearTokenWatch()

    const tryConnect = (): void => {
      socket.connect(id, handleEnvelope)
      socket.retryPendingConnect()
    }

    if (auth.token) {
      tryConnect()
      return
    }

    stopTokenWatch = watch(
      () => auth.token,
      (token) => {
        if (token && gameId.value === id) {
          tryConnect()
          clearTokenWatch()
        }
      },
    )
  }

  function disconnect(): void {
    reset()
  }

  function maybeSendLocalState(): void {
    const id = gameId.value
    if (!id || !socket.connected.value || matchEnded.value) return

    const session = useGameSessionStore()
    if (!session.active || !session.engine) return

    const now = performance.now()
    if (now - lastSentAt < SEND_INTERVAL_MS) return
    lastSentAt = now

    const alive = !session.gameOver && session.matchEndKind === 'playing'
    socket.sendPlayerState(id, {
      score: session.score,
      lines: session.lines,
      level: session.level,
      alive,
      board: encodeLockedMatrixBase64(session.engine.matrixRef),
    })
  }

  return {
    gameId,
    players,
    opponents,
    opponentList,
    matchError,
    matchEnded,
    matchResults,
    connected,
    bootstrapFailed,
    bootstrap,
    getSharedSeed,
    ensurePlayersRoster,
    seedFromMatchStart,
    connect,
    disconnect,
    maybeSendLocalState,
    notifyLocalElimination,
    applyOpponentState,
    applyMatchEnded,
  }
})
