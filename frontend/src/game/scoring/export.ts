import type { MatchRecordV1, ScoreBreakdown, ScoreSnapshot } from '@/game/scoring/types'

export type MatchRecordInput = {
  runId: string
  seed: number
  variation: string
  playerCount: number
  startedAt: string
  endedAt?: string
  final: ScoreSnapshot
  events: readonly ScoreBreakdown[]
}

/** Build a JSON-serializable match record for a future backend API. */
export function toMatchRecordV1(input: MatchRecordInput): MatchRecordV1 {
  return {
    schemaVersion: 1,
    runId: input.runId,
    seed: input.seed,
    variation: input.variation,
    playerCount: input.playerCount,
    startedAt: input.startedAt,
    endedAt: input.endedAt,
    final: { ...input.final },
    events: input.events.map((e) => ({ ...e })),
  }
}
