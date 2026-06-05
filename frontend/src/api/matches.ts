import { apiFetch } from '@/api/client'
import type { MatchDetail, MatchResponse } from '@/types/api'

export async function getMatch(gameId: string): Promise<MatchDetail | null> {
  try {
    const res = await apiFetch<MatchResponse | null>(`/matches/${gameId}`, {
      method: 'GET',
      allowStatuses: [404],
    })
    return res?.match ?? null
  } catch {
    return null
  }
}
