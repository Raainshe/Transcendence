import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import * as lobbiesApi from '@/api/lobbies'
import { ApiError } from '@/api/client'
import {
  WS_TYPE_ERROR,
  WS_TYPE_LOBBY_CLOSED,
  WS_TYPE_LOBBY_UPDATED,
  WS_TYPE_MATCH_START,
  type LobbyClosedPayload,
  type MatchStartPayload,
  type WsEnvelope,
} from '@/composables/useLobbySocket'
import { i18n } from '@/i18n'
import { useAuthStore } from '@/stores/auth'
import type { LobbyDetail, MatchPlayerView, StartLobbyResult } from '@/types/api'

const MP_SEED_PREFIX = 'mp:'
const MP_PLAYERS_PREFIX = 'mp-players:'

export function stashMultiplayerSeed(gameId: string, seed: number): void {
  if (typeof sessionStorage === 'undefined') return
  try {
    sessionStorage.setItem(`${MP_SEED_PREFIX}${gameId}`, String(seed))
  } catch {
    /* quota / private mode */
  }
}

export function readMultiplayerSeed(gameId: string): number | null {
  if (typeof sessionStorage === 'undefined') return null
  try {
    const raw = sessionStorage.getItem(`${MP_SEED_PREFIX}${gameId}`)
    if (!raw) return null
    const n = Number(raw)
    return Number.isFinite(n) && n > 0 ? n : null
  } catch {
    return null
  }
}

export function stashMatchPlayers(gameId: string, players: MatchPlayerView[]): void {
  if (typeof sessionStorage === 'undefined' || players.length === 0) return
  try {
    sessionStorage.setItem(`${MP_PLAYERS_PREFIX}${gameId}`, JSON.stringify(players))
  } catch {
    /* quota / private mode */
  }
}

export function readMatchPlayers(gameId: string): MatchPlayerView[] {
  if (typeof sessionStorage === 'undefined') return []
  try {
    const raw = sessionStorage.getItem(`${MP_PLAYERS_PREFIX}${gameId}`)
    if (!raw) return []
    const parsed = JSON.parse(raw) as MatchPlayerView[]
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export const useLobbyStore = defineStore('lobby', () => {
  const auth = useAuthStore()

  const lobby = ref<LobbyDetail | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const actionBusy = ref(false)
  const closedReason = ref<LobbyClosedPayload['reason'] | null>(null)

  const isHost = computed(() => {
    if (!lobby.value || !auth.user) return false
    return lobby.value.host_user_id === auth.user.id
  })

  const myMember = computed(() => {
    if (!lobby.value || !auth.user) return null
    return lobby.value.members.find((m) => m.user_id === auth.user!.id) ?? null
  })

  const allReady = computed(() => {
    if (!lobby.value || lobby.value.members.length === 0) return false
    return lobby.value.members.every((m) => m.is_ready)
  })

  const canStart = computed(() => {
    if (!lobby.value || lobby.value.status !== 'waiting') return false
    if (!isHost.value) return false
    if (lobby.value.members.length < 2) return false
    return allReady.value
  })

  function mapError(err: unknown): string {
    if (err instanceof ApiError) {
      if (err.status === 404) {
        return i18n.global.t('lobby.errors.notFound')
      }
      const known: Record<string, string> = {
        'max_players must be between 2 and 4': 'lobby.errors.invalidMaxPlayers',
        'only the lobby host can start the match': 'lobby.errors.notHost',
        'lobby is not joinable': 'lobby.errors.notJoinable',
        'lobby is full': 'lobby.errors.lobbyFull',
        'user is already in this lobby': 'lobby.errors.alreadyInLobby',
        'user is already in another waiting lobby': 'lobby.errors.inAnotherLobby',
        'at least 2 players are required to start': 'lobby.errors.notEnoughPlayers',
        'all players must be ready before starting': 'lobby.errors.notAllReady',
        'user is not a lobby member': 'lobby.errors.notMember',
        'lobby not found': 'lobby.errors.notFound',
        'invite_code is required': 'lobby.errors.codeRequired',
      }
      const mapped = known[err.message.toLowerCase()]
      if (mapped) return i18n.global.t(mapped)
      if (err.message) return err.message
    }
    return i18n.global.t('lobby.errors.generic')
  }

  function reset(): void {
    lobby.value = null
    loading.value = false
    error.value = null
    actionBusy.value = false
    closedReason.value = null
  }

  async function load(id: string): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      const res = await lobbiesApi.getLobby(id)
      if (!res) {
        error.value = i18n.global.t('lobby.errors.notFound')
        lobby.value = null
        return false
      }
      lobby.value = res.lobby
      return true
    } catch (err) {
      error.value = mapError(err)
      lobby.value = null
      return false
    } finally {
      loading.value = false
    }
  }

  async function toggleReady(): Promise<void> {
    if (!lobby.value || actionBusy.value) return
    const member = myMember.value
    if (!member) return
    actionBusy.value = true
    error.value = null
    try {
      const res = await lobbiesApi.setReady(lobby.value.id, { ready: !member.is_ready })
      lobby.value = res.lobby
    } catch (err) {
      error.value = mapError(err)
    } finally {
      actionBusy.value = false
    }
  }

  async function start(): Promise<StartLobbyResult | null> {
    if (!lobby.value || actionBusy.value || !canStart.value) return null
    actionBusy.value = true
    error.value = null
    try {
      const result = await lobbiesApi.startLobby(lobby.value.id)
      stashMultiplayerSeed(result.game_id, result.shared_seed)
      stashMatchPlayers(result.game_id, result.players)
      return result
    } catch (err) {
      error.value = mapError(err)
      return null
    } finally {
      actionBusy.value = false
    }
  }

  async function leave(): Promise<void> {
    if (!lobby.value || actionBusy.value) return
    actionBusy.value = true
    error.value = null
    try {
      await lobbiesApi.leaveLobby(lobby.value.id)
      reset()
    } catch (err) {
      error.value = mapError(err)
    } finally {
      actionBusy.value = false
    }
  }

  function applyWsEnvelope(env: WsEnvelope): StartLobbyResult | null {
    switch (env.type) {
      case WS_TYPE_LOBBY_UPDATED: {
        const payload = env.payload as { lobby?: LobbyDetail } | undefined
        if (payload?.lobby) {
          lobby.value = payload.lobby
        }
        return null
      }
      case WS_TYPE_LOBBY_CLOSED: {
        const payload = env.payload as LobbyClosedPayload | undefined
        closedReason.value = payload?.reason ?? 'host_left'
        lobby.value = null
        return null
      }
      case WS_TYPE_MATCH_START: {
        const payload = env.payload as MatchStartPayload | undefined
        if (!payload?.game_id) return null
        stashMultiplayerSeed(payload.game_id, payload.shared_seed)
        stashMatchPlayers(payload.game_id, payload.players)
        return {
          game_id: payload.game_id,
          shared_seed: payload.shared_seed,
          players: payload.players,
        }
      }
      case WS_TYPE_ERROR: {
        const payload = env.payload as { error?: string } | undefined
        const msg = payload?.error?.toLowerCase() ?? ''
        if (msg.includes('not a lobby member') || msg.includes('invalid lobby')) {
          error.value = i18n.global.t('lobby.errors.notFound')
        } else if (payload?.error) {
          error.value = mapError(new ApiError(400, payload.error))
        }
        return null
      }
      default:
        return null
    }
  }

  return {
    lobby,
    loading,
    error,
    actionBusy,
    closedReason,
    isHost,
    myMember,
    allReady,
    canStart,
    reset,
    load,
    toggleReady,
    start,
    leave,
    applyWsEnvelope,
    mapError,
  }
})
