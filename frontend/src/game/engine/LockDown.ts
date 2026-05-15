/**
 * Extended Placement lock-down policy (Guideline §5.7).
 *
 * Behavior summary:
 * - When a piece can no longer fall, a 500 ms timer (`LOCK_DELAY_MS`) starts.
 * - On every successful move or rotation while the timer is active, the
 *   timer resets to 500 ms.
 * - The number of such resets is capped at 15 per piece (`MAX_LOCK_RESETS`).
 *   Past the cap, moves/rotations no longer refresh the timer and the piece
 *   locks on the next surface it touches.
 * - "If the tetrimino falls one row below the lowest row yet reached, this
 *   counter is reset." — i.e. dropping to a brand-new lowest row refills the
 *   15-reset budget. Going *up* (e.g. a wall kick) does NOT refill.
 * - If a rotation/move lifts the piece off the surface so it can fall again,
 *   the timer pauses but the move counter does NOT reset until the new
 *   lowest row is reached.
 *
 * Designed as a self-contained class so slice 8 can introduce sibling
 * `Infinite` and `Classic` policies behind the same interface.
 */

import {
  LOCK_DELAY_MS,
  type LockDownSnapshot,
  MAX_LOCK_RESETS,
} from '@/game/types'

export type LockDownConfig = {
  lockDelayMs?: number
  maxResets?: number
}

export class LockDown {
  private readonly lockDelayMs: number
  private readonly maxResets: number

  private timerMs: number
  private movesLeft: number
  /**
   * Lowest bottom-y the current piece has reached (in guideline y+up coords).
   * Initialized to `+Infinity` so the first `onFell`/`reset` always sets it.
   */
  private lowestBottomY: number
  private active: boolean

  constructor(config: LockDownConfig = {}) {
    this.lockDelayMs = config.lockDelayMs ?? LOCK_DELAY_MS
    this.maxResets = config.maxResets ?? MAX_LOCK_RESETS
    this.timerMs = this.lockDelayMs
    this.movesLeft = this.maxResets
    this.lowestBottomY = Number.POSITIVE_INFINITY
    this.active = false
  }

  /** Resets the policy for a freshly spawned piece. */
  reset(bottomY: number): void {
    this.timerMs = this.lockDelayMs
    this.movesLeft = this.maxResets
    this.lowestBottomY = bottomY
    this.active = false
  }

  /** Called when the piece can no longer fall. Starts the timer if not already active. */
  onLanded(): void {
    if (!this.active) {
      this.active = true
      this.timerMs = this.lockDelayMs
    }
  }

  /**
   * Called when the piece can fall again (e.g. a rotation kicked it up off
   * the surface). The timer pauses; the move counter is left as-is per §5.7.
   */
  onLifted(): void {
    this.active = false
  }

  /**
   * Called after a successful move or rotation. Returns `true` if the action
   * consumed a reset (i.e. the timer was refreshed). Returns `false` if the
   * piece is not yet on a surface (no-op) or the reset budget is exhausted.
   */
  onMovedOrRotated(): boolean {
    if (!this.active) return false
    if (this.movesLeft <= 0) return false
    this.timerMs = this.lockDelayMs
    this.movesLeft -= 1
    return true
  }

  /**
   * Called after gravity (or soft drop) moves the piece down. If the new
   * bottom-y is below the lowest yet reached, refills the move budget
   * (§5.7 "falls one row below the lowest row yet reached").
   */
  onFell(newBottomY: number): void {
    if (newBottomY < this.lowestBottomY) {
      this.lowestBottomY = newBottomY
      this.movesLeft = this.maxResets
    }
  }

  /**
   * Advances the timer by `dtMs` if active. Returns `true` exactly once, on
   * the tick that the timer crosses zero — that's the signal for the engine
   * to lock the piece. Subsequent calls after expiry are no-ops (return
   * `false`) until `reset()` is called for a new piece.
   */
  tick(dtMs: number): boolean {
    if (!this.active) return false
    if (this.timerMs <= 0) return false
    this.timerMs -= dtMs
    if (this.timerMs <= 0) {
      this.timerMs = 0
      return true
    }
    return false
  }

  snapshot(): LockDownSnapshot {
    return {
      active: this.active,
      timerMs: this.timerMs,
      movesLeft: this.movesLeft,
      lowestBottomY: this.lowestBottomY,
    }
  }
}
