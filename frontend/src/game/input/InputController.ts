/**
 * Pure-TS input controller: turns logical commands (`press` / `release`) into
 * `Engine` method calls, applying Delayed Auto Shift (DAS) and Auto-Repeat
 * Rate (ARR) to horizontal moves. No DOM dependencies — a `KeyboardAdapter`
 * (or a future gamepad / touch adapter) feeds it.
 *
 * Time model: callers drive the controller by calling `update(dtMs)` once per
 * frame, before the engine's own `update(dt)`. The controller owns no
 * wall-clock state outside that call — fully deterministic and testable.
 *
 * Behavior:
 * - Move L/R: fire one move immediately on press; after `dasMs` of being
 *   held, auto-repeat at `arrMs` cadence. `arrMs === 0` switches to a "burst
 *   to wall" mode (capped by `MAX_ARR_BURST`).
 * - Conflict (both directions held): most-recent direction wins. On release
 *   of the active direction, fall back to the still-held opposing direction
 *   with a fresh DAS charge plus an immediate move.
 * - Rotate CW/CCW, Hard drop: single-shot per press, no auto-repeat.
 * - Soft drop: held — toggles `engine.softDrop(true|false)`. The engine's
 *   own 20× gravity multiplier (slice 3) handles the timing.
 * - `press` / `release` are idempotent so the adapter can forward `keydown`
 *   repeats blindly.
 * - `releaseAll()` clears every held command + soft drop, intended for use
 *   on window blur / `visibilitychange === 'hidden'`.
 */

import {
  DEFAULT_ARR_MS,
  DEFAULT_DAS_MS,
  InputCommand,
  type InputConfig,
  MAX_ARR_BURST,
} from '@/game/types'

/**
 * Minimal interface the controller needs from the `Engine`. Defined here
 * (rather than imported from `Engine`) so tests can drop in a tiny stub
 * without constructing a full engine + matrix + bag.
 */
export interface EngineLike {
  moveLeft(): boolean
  moveRight(): boolean
  rotateCW(): boolean
  rotateCCW(): boolean
  softDrop(active: boolean): void
  hardDrop(): boolean
  hold(): boolean
}

type DirState = {
  held: boolean
  dasTimerMs: number
  arrTimerMs: number
  charged: boolean
}

function makeDirState(): DirState {
  return { held: false, dasTimerMs: 0, arrTimerMs: 0, charged: false }
}

export class InputController {
  private readonly engine: EngineLike
  private readonly dasMs: number
  private readonly arrMs: number
  private readonly isInputBlocked: () => boolean

  private readonly left: DirState = makeDirState()
  private readonly right: DirState = makeDirState()
  private activeDir: 'left' | 'right' | null = null

  private rotateCwHeld = false
  private rotateCcwHeld = false
  private hardDropHeld = false
  private holdHeld = false
  private softDropHeld = false

  constructor(engine: EngineLike, config: InputConfig = {}) {
    this.engine = engine
    this.dasMs = config.dasMs ?? DEFAULT_DAS_MS
    this.arrMs = config.arrMs ?? DEFAULT_ARR_MS
    this.isInputBlocked = config.isInputBlocked ?? (() => false)
  }

  // -------------------------------------------------------------------------
  // Press / release
  // -------------------------------------------------------------------------

  press(cmd: InputCommand): void {
    if (this.isInputBlocked()) return
    switch (cmd) {
      case InputCommand.MoveLeft:
        if (this.left.held) return
        this.left.held = true
        this.activate('left')
        return
      case InputCommand.MoveRight:
        if (this.right.held) return
        this.right.held = true
        this.activate('right')
        return
      case InputCommand.RotateCW:
        if (this.rotateCwHeld) return
        this.rotateCwHeld = true
        this.engine.rotateCW()
        return
      case InputCommand.RotateCCW:
        if (this.rotateCcwHeld) return
        this.rotateCcwHeld = true
        this.engine.rotateCCW()
        return
      case InputCommand.HardDrop:
        if (this.hardDropHeld) return
        this.hardDropHeld = true
        this.engine.hardDrop()
        return
      case InputCommand.Hold:
        if (this.holdHeld) return
        this.holdHeld = true
        this.engine.hold()
        return
      case InputCommand.SoftDrop:
        if (this.softDropHeld) return
        this.softDropHeld = true
        this.engine.softDrop(true)
        return
    }
  }

  release(cmd: InputCommand): void {
    switch (cmd) {
      case InputCommand.MoveLeft:
        if (!this.left.held) return
        this.left.held = false
        this.left.charged = false
        if (this.activeDir === 'left') {
          if (this.right.held) this.activate('right')
          else this.activeDir = null
        }
        return
      case InputCommand.MoveRight:
        if (!this.right.held) return
        this.right.held = false
        this.right.charged = false
        if (this.activeDir === 'right') {
          if (this.left.held) this.activate('left')
          else this.activeDir = null
        }
        return
      case InputCommand.RotateCW:
        this.rotateCwHeld = false
        return
      case InputCommand.RotateCCW:
        this.rotateCcwHeld = false
        return
      case InputCommand.HardDrop:
        this.hardDropHeld = false
        return
      case InputCommand.Hold:
        this.holdHeld = false
        return
      case InputCommand.SoftDrop:
        if (!this.softDropHeld) return
        this.softDropHeld = false
        this.engine.softDrop(false)
        return
    }
  }

  releaseAll(): void {
    this.left.held = false
    this.left.charged = false
    this.right.held = false
    this.right.charged = false
    this.activeDir = null
    this.rotateCwHeld = false
    this.rotateCcwHeld = false
    this.hardDropHeld = false
    this.holdHeld = false
    if (this.softDropHeld) {
      this.softDropHeld = false
      this.engine.softDrop(false)
    }
  }

  // -------------------------------------------------------------------------
  // Per-frame DAS / ARR
  // -------------------------------------------------------------------------

  update(dtMs: number): void {
    if (this.isInputBlocked()) return
    if (this.activeDir === null) return
    if (dtMs <= 0) return

    const state = this.activeDir === 'left' ? this.left : this.right
    let remaining = dtMs

    if (!state.charged) {
      state.dasTimerMs -= remaining
      if (state.dasTimerMs > 0) return
      // DAS expired this tick. Carry the leftover into the ARR phase so a
      // single very-large `update(dt)` produces the right number of moves.
      const leftover = -state.dasTimerMs
      state.dasTimerMs = 0
      state.charged = true
      state.arrTimerMs = 0
      remaining = leftover
    }

    if (this.arrMs === 0) {
      // Infinite-ARR: empty the auto-shift in one shot, bounded by the
      // safety cap so a misbehaving engine can't hang the frame.
      for (let i = 0; i < MAX_ARR_BURST; i++) {
        if (!this.fireActiveMove()) break
      }
      state.arrTimerMs = 0
      return
    }

    state.arrTimerMs -= remaining
    while (state.arrTimerMs <= 0) {
      if (!this.fireActiveMove()) {
        // Wall (or other rejection). Wait a full ARR window before retrying;
        // this avoids spamming engine.move* every frame against a wall.
        state.arrTimerMs = this.arrMs
        break
      }
      state.arrTimerMs += this.arrMs
    }
  }

  // -------------------------------------------------------------------------
  // Internals
  // -------------------------------------------------------------------------

  /**
   * Switches the active direction to `dir`, refreshes its DAS/ARR state, and
   * fires one immediate move (the "tap" behavior expected on both initial
   * press and on opposing-direction release-fallback).
   */
  private activate(dir: 'left' | 'right'): void {
    this.activeDir = dir
    const s = dir === 'left' ? this.left : this.right
    s.dasTimerMs = this.dasMs
    s.arrTimerMs = this.arrMs
    s.charged = false
    this.fireActiveMove()
  }

  private fireActiveMove(): boolean {
    return this.activeDir === 'left' ? this.engine.moveLeft() : this.engine.moveRight()
  }
}
