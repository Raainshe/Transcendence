import { apiFetch } from '@/api/client'
import type {
  CreateGamePayload,
  CreateGameResponse,
  GameDetailResponse,
  GamesListResponse,
  LeaderboardEntry,
} from '@/types/api'
import type { MatchRecordV1 } from '@/game/scoring/types'

export function matchRecordToCreateGame(
  record: MatchRecordV1,
  isWinner: boolean,
): CreateGamePayload | null {
  if (!record.endedAt) return null
  return {
    mode: record.variation,
    score: record.final.score,
    lines_cleared: record.final.lines,
    level_reached: record.final.level,
    started_at: record.startedAt,
    finished_at: record.endedAt,
    is_winner: isWinner,
  }
}

export function createGame(payload: CreateGamePayload): Promise<CreateGameResponse> {
  return apiFetch<CreateGameResponse>('/games', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export type ListGamesParams = {
  userId?: string
  limit?: number
  offset?: number
}

export function listGames(params: ListGamesParams = {}): Promise<GamesListResponse> {
  const search = new URLSearchParams()
  if (params.userId) search.set('user_id', params.userId)
  if (params.limit !== undefined && params.limit > 0) search.set('limit', String(params.limit))
  if (params.offset !== undefined && params.offset > 0) search.set('offset', String(params.offset))
  const query = search.toString()
  return apiFetch<GamesListResponse>(`/games${query ? `?${query}` : ''}`, {
    method: 'GET',
    auth: false,
  })
}

export function getGame(gameId: string): Promise<GameDetailResponse> {
  return apiFetch<GameDetailResponse>(`/games/${gameId}`, {
    method: 'GET',
    auth: false,
  })
}

export function getLeaderboard(limit = 50): Promise<LeaderboardEntry[]> {
  const params = new URLSearchParams()
  if (limit > 0) params.set('limit', String(limit))
  const query = params.toString()
  return apiFetch<LeaderboardEntry[]>(`/leaderboard${query ? `?${query}` : ''}`, {
    method: 'GET',
    auth: false,
  })
}
