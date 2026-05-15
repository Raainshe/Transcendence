/**
 * Serializable scoring types for HUD, engine events, and future backend POST bodies.
 * Plain objects only — no class instances, no Vue refs.
 */

/** Line clear classification from row count (§8). */
export type LineClearKind = 'none' | 'single' | 'double' | 'triple' | 'tetris'

/** T-Spin detection result at lock (§9). */
export type TSpinKind = 'none' | 'mini' | 'full'

/** Why points were awarded — stable for JSON / API. */
export type ScoreReason = 'lineClear' | 'softDrop' | 'hardDrop' | 'tSpinNoLines'

/**
 * One scoring action (append-only ledger entry).
 * `sequence` is monotonic per match for backend ordering.
 */
export type ScoreBreakdown = {
  sequence: number
  reason: ScoreReason
  level: number
  linesCleared: number
  lineClearKind: LineClearKind
  tSpinKind: TSpinKind
  basePoints: number
  /** 1 or 1.5 (Back-to-Back §8). */
  backToBackMultiplier: number
  pointsAwarded: number
  /** B2B chain length after this action (0 = inactive). */
  backToBackChain: number
}

/** Running totals exposed on engine state and HUD. */
export type ScoreSnapshot = {
  score: number
  level: number
  lines: number
  /** True when the next difficult clear earns ×1.5. */
  backToBackActive: boolean
  /** Longest B2B chain this match (for stats / backend). */
  backToBackCount: number
}

/** Future `POST /matches` body shape — versioned. */
export type MatchRecordV1 = {
  schemaVersion: 1
  runId: string
  seed: number
  variation: string
  playerCount: number
  startedAt: string
  endedAt?: string
  final: ScoreSnapshot
  events: ScoreBreakdown[]
}
