import { apiFetch } from '@/api/client'
import type { CreateGamePayload, CreateGameResponse, LeaderboardEntry } from '@/types/api'
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

export function getLeaderboard(limit = 50): Promise<LeaderboardEntry[]> {
  const params = new URLSearchParams()
  if (limit > 0) params.set('limit', String(limit))
  const query = params.toString()
  return apiFetch<LeaderboardEntry[]>(`/leaderboard${query ? `?${query}` : ''}`, {
    method: 'GET',
    auth: false,
  })
}
