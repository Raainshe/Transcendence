/**
 * Tetris Guideline scoring (§8): line clears, T-Spins, soft/hard drop, Back-to-Back.
 * Pure functions — no engine coupling.
 */

import type { LineClearKind, ScoreBreakdown, TSpinKind } from '@/game/scoring/types'

/** Standard line clear base points per level (§8). */
export const LINE_CLEAR_BASE: Readonly<Record<Exclude<LineClearKind, 'none'>, number>> = {
  single: 100,
  double: 300,
  triple: 500,
  tetris: 800,
}

/** T-Spin line clear base points per level (§8). */
export const TSPIN_LINE_BASE: Readonly<
  Record<1 | 2 | 3, { full: number; mini: number }>
> = {
  1: { full: 800, mini: 200 },
  2: { full: 1200, mini: 400 },
  3: { full: 1600, mini: 600 },
}

/** T-Spin with zero lines cleared (§8). */
export const TSPIN_NO_LINES_BASE = { full: 400, mini: 100 } as const

export const SOFT_DROP_POINTS_PER_CELL = 1
export const HARD_DROP_POINTS_PER_CELL = 2
export const BACK_TO_BACK_MULTIPLIER = 1.5

export function lineClearKindFromCount(lines: number): LineClearKind {
  if (lines <= 0) return 'none'
  if (lines === 1) return 'single'
  if (lines === 2) return 'double'
  if (lines === 3) return 'triple'
  return 'tetris'
}

/**
 * Difficult clears that maintain / advance Back-to-Back (§8):
 * Tetris or any T-Spin that clears at least one line.
 */
export function isDifficultClear(
  lineClearKind: LineClearKind,
  tSpinKind: TSpinKind,
  linesCleared: number,
): boolean {
  if (lineClearKind === 'tetris') return true
  if (linesCleared > 0 && tSpinKind !== 'none') return true
  return false
}

function basePointsForClear(
  lineClearKind: LineClearKind,
  tSpinKind: TSpinKind,
  linesCleared: number,
): number {
  if (linesCleared === 0) {
    if (tSpinKind === 'full') return TSPIN_NO_LINES_BASE.full
    if (tSpinKind === 'mini') return TSPIN_NO_LINES_BASE.mini
    return 0
  }

  if (tSpinKind !== 'none' && linesCleared >= 1 && linesCleared <= 3) {
    const row = TSPIN_LINE_BASE[linesCleared as 1 | 2 | 3]
    return tSpinKind === 'full' ? row.full : row.mini
  }

  if (lineClearKind === 'none') return 0
  return LINE_CLEAR_BASE[lineClearKind]
}

export type ScoreLineClearInput = {
  sequence: number
  level: number
  linesCleared: number
  lineClearKind: LineClearKind
  tSpinKind: TSpinKind
  backToBackActive: boolean
  backToBackChain: number
}

export type ScoreLineClearResult = {
  breakdown: ScoreBreakdown | null
  nextBackToBackActive: boolean
  nextBackToBackChain: number
}

export function scoreLineClear(input: ScoreLineClearInput): ScoreLineClearResult {
  const { level, linesCleared, lineClearKind, tSpinKind } = input
  const base = basePointsForClear(lineClearKind, tSpinKind, linesCleared)

  if (base === 0 && linesCleared === 0 && tSpinKind === 'none') {
    return {
      breakdown: null,
      nextBackToBackActive: false,
      nextBackToBackChain: 0,
    }
  }

  const difficult = isDifficultClear(lineClearKind, tSpinKind, linesCleared)
  let multiplier = 1
  let nextChain = 0
  let nextActive = false

  if (difficult) {
    if (input.backToBackActive) {
      multiplier = BACK_TO_BACK_MULTIPLIER
      nextChain = input.backToBackChain + 1
    } else {
      nextChain = 1
    }
    nextActive = true
  } else if (linesCleared > 0) {
    nextActive = false
    nextChain = 0
  } else {
    nextActive = input.backToBackActive
    nextChain = input.backToBackChain
  }

  const pointsAwarded = Math.floor(base * level * multiplier)
  const reason =
    linesCleared === 0 && tSpinKind !== 'none' ? 'tSpinNoLines' : 'lineClear'

  const breakdown: ScoreBreakdown = {
    sequence: input.sequence,
    reason,
    level,
    linesCleared,
    lineClearKind,
    tSpinKind,
    basePoints: base,
    backToBackMultiplier: multiplier,
    pointsAwarded,
    backToBackChain: nextChain,
  }

  return {
    breakdown,
    nextBackToBackActive: nextActive,
    nextBackToBackChain: nextChain,
  }
}

export function scoreDropCells(
  sequence: number,
  cells: number,
  kind: 'soft' | 'hard',
  level: number,
): ScoreBreakdown | null {
  if (cells <= 0) return null
  const perCell = kind === 'soft' ? SOFT_DROP_POINTS_PER_CELL : HARD_DROP_POINTS_PER_CELL
  const basePoints = perCell * cells
  const pointsAwarded = basePoints * level
  return {
    sequence,
    reason: kind === 'soft' ? 'softDrop' : 'hardDrop',
    level,
    linesCleared: 0,
    lineClearKind: 'none',
    tSpinKind: 'none',
    basePoints,
    backToBackMultiplier: 1,
    pointsAwarded,
    backToBackChain: 0,
  }
}
