/**
 * Gravity timing — converts the current level into a fall interval in
 * milliseconds per cell.
 *
 * Formula from Guideline §7:
 *
 *     time_per_line(level) = (0.8 - ((level - 1) * 0.007)) ^ (level - 1)  seconds
 *
 * which yields 1.000s at level 1, 0.180s at level 9, 0.028s at level 15, etc.
 * The full table only goes up to level 15 in the Fixed Goal System; past that
 * the formula begins to blow up (the base goes negative around level ~115).
 * We clamp to `[1, MAX_LEVEL]` here so callers can pass raw values safely.
 *
 * Soft drop is 20x faster than natural gravity (§5.5).
 */

import { MAX_LEVEL, SOFT_DROP_MULTIPLIER } from '@/game/types'

/**
 * Milliseconds per gravity step (one cell of fall) at the given level.
 * `level` is clamped to `[1, MAX_LEVEL]` and floored.
 */
export function msPerCell(level: number): number {
  const clamped = Math.max(1, Math.min(MAX_LEVEL, Math.floor(level)))
  const secPerCell = Math.pow(0.8 - (clamped - 1) * 0.007, clamped - 1)
  return secPerCell * 1000
}

/** Soft-drop fall interval — exactly `msPerCell(level) / SOFT_DROP_MULTIPLIER`. */
export function softDropMsPerCell(level: number): number {
  return msPerCell(level) / SOFT_DROP_MULTIPLIER
}
