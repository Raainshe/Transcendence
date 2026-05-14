/**
 * Core Tetris engine — pure TypeScript, framework-agnostic.
 *
 * Drives the 8-phase loop from §A1.1 (Generation, Falling, Lock, Pattern,
 * Iterate, Animate, Eliminate, Completion). In practice only Generation,
 * Falling, and Lock have meaningful duration; the rest collapse into a
 * synchronous "commit" inside `commitLock()`.
 *
 * Time model: callers tick the engine with `update(dtMs)`. The engine
 * deterministically resolves phase transitions across that interval so a
 * single `update(1000)` is equivalent to ten `update(100)` calls (modulo
 * floating-point). Inputs (`moveLeft`, `rotateCW`, `hardDrop`, ...) take
 * effect immediately and may transition phases (Falling ⇄ Lock) on their own.
 *
 * Slice 3–6: gravity, lock-down, line clear, level-up, block-out, **Hold** (§2.5),
 * ghost rendering (UI). Scoring, T-spin, variants, Lock-out / Top-out land later.
 */

import { Bag } from '@/game/engine/Bag'
import { msPerCell, softDropMsPerCell } from '@/game/engine/FallSpeed'
import { LockDown } from '@/game/engine/LockDown'
import { Matrix } from '@/game/engine/Matrix'
import { type RotationDirection, tryRotate as srsTryRotate } from '@/game/engine/SRS'
import { getOccupiedCells, spawn, translate } from '@/game/engine/Tetrimino'
import {
  type EngineConfig,
  type EngineEvent,
  EnginePhase,
  type EngineState,
  GENERATION_DELAY_MS,
  type GameOverReason,
  LINES_PER_LEVEL,
  MAX_LEVEL,
  type PieceType,
  type Tetrimino,
} from '@/game/types'

const DEFAULT_NEXT_QUEUE_SIZE = 6

/**
 * Safety cap on while-loop iterations inside `update()`. Real games are at
 * most a handful of phase transitions per frame; 10 000 is a generous bound
 * that protects against an accidentally infinite loop without affecting
 * legitimate gameplay.
 */
const MAX_UPDATE_ITERATIONS = 10_000

export class Engine {
  private readonly matrix: Matrix
  private readonly bag: Bag
  private readonly lockDown: LockDown
  private readonly generationDelayMs: number
  private readonly nextQueueSize: number
  private readonly startLevel: number

  private phase: EnginePhase
  private currentPiece: Tetrimino | null
  private level: number
  private lines: number
  private goal: number
  private softDropActive: boolean
  private gravityAccumMs: number
  private generationTimerMs: number
  private gameOverReason: GameOverReason | undefined
  private events: EngineEvent[]

  private holdSlot: PieceType | null = null
  /** True after a successful Hold until the current piece locks (§2.5.3). */
  private holdUsedForCurrentPiece = false

  constructor(config: EngineConfig = {}) {
    this.matrix = new Matrix()
    this.bag = new Bag(config.seed)
    this.lockDown = new LockDown({
      lockDelayMs: config.lockDelayMs,
      maxResets: config.maxLockResets,
    })

    this.generationDelayMs = config.generationDelayMs ?? GENERATION_DELAY_MS
    this.nextQueueSize = config.nextQueueSize ?? DEFAULT_NEXT_QUEUE_SIZE
    this.startLevel = clampLevel(config.startLevel ?? 1)

    this.phase = EnginePhase.Generation
    this.currentPiece = null
    this.level = this.startLevel
    this.lines = 0
    this.goal = LINES_PER_LEVEL
    this.softDropActive = false
    this.gravityAccumMs = 0
    // The very first piece is also gated by the generation delay so the
    // engine behaves uniformly tick-to-tick. Tests call `update(200)` to
    // arm the first spawn.
    this.generationTimerMs = this.generationDelayMs
    this.gameOverReason = undefined
    this.events = []
  }

  // -------------------------------------------------------------------------
  // Public introspection
  // -------------------------------------------------------------------------

  get state(): EngineState {
    return {
      phase: this.phase,
      currentPiece: this.currentPiece,
      nextQueue: this.bag.peek(this.nextQueueSize) as readonly PieceType[],
      holdPiece: this.holdSlot,
      canHold: this.computeCanHold(),
      level: this.level,
      lines: this.lines,
      goal: this.goal,
      softDropActive: this.softDropActive,
      generationTimerMs: this.generationTimerMs,
      lockDown: this.lockDown.snapshot(),
      gameOver: this.phase === EnginePhase.GameOver,
      gameOverReason: this.gameOverReason,
    }
  }

  /** Direct access to the underlying matrix. Treat as read-only. */
  get matrixRef(): Matrix {
    return this.matrix
  }

  drainEvents(): EngineEvent[] {
    const out = this.events
    this.events = []
    return out
  }

  // -------------------------------------------------------------------------
  // Time-driven loop
  // -------------------------------------------------------------------------

  update(dtMs: number): void {
    if (this.phase === EnginePhase.GameOver) return
    if (dtMs <= 0) return

    let remaining = dtMs
    for (let i = 0; i < MAX_UPDATE_ITERATIONS && remaining > 0; i++) {
      // Cast widens the type back from the outer-guard narrowing — phase
      // can become `GameOver` mid-loop via `spawnNext()` -> `setGameOver()`.
      const phaseBefore = this.phase as EnginePhase
      if (phaseBefore === EnginePhase.GameOver) return

      if (phaseBefore === EnginePhase.Generation) {
        const consume = Math.min(remaining, this.generationTimerMs)
        this.generationTimerMs -= consume
        remaining -= consume
        if (this.generationTimerMs <= 0) {
          this.generationTimerMs = 0
          this.spawnNext()
        } else {
          // Still waiting for the next piece and we exhausted dt.
          break
        }
        continue
      }

      if (phaseBefore === EnginePhase.Falling) {
        // Gravity is bucketed in milliseconds. Each step of `fallInterval`
        // moves the piece down one row. Soft drop swaps in the faster rate.
        this.gravityAccumMs += remaining
        remaining = 0
        const fallInterval = this.fallInterval()
        // Guard against pathological intervals at level 15 still being
        // > 0; if fallInterval is 0 we'd loop forever.
        const safeInterval = fallInterval > 0 ? fallInterval : 1
        let safety = MAX_UPDATE_ITERATIONS
        while (
          this.phase === EnginePhase.Falling &&
          this.gravityAccumMs >= safeInterval &&
          safety > 0
        ) {
          this.gravityAccumMs -= safeInterval
          this.applyGravityStep()
          safety -= 1
        }
        // If we transitioned to Lock, abandon the leftover gravity bucket —
        // the lock-down timer is a fresh 500 ms window, not a continuation.
        if (this.phase !== EnginePhase.Falling) {
          this.gravityAccumMs = 0
        }
        continue
      }

      if (phaseBefore === EnginePhase.Lock) {
        const snap = this.lockDown.snapshot()
        const consume = Math.min(remaining, snap.timerMs)
        remaining -= consume
        const expired = this.lockDown.tick(consume)
        if (expired) {
          this.commitLock()
          continue
        }
        // Timer not yet expired and we used all the time we were going to
        // use in Lock. Bail out.
        break
      }
    }
  }

  // -------------------------------------------------------------------------
  // Inputs
  // -------------------------------------------------------------------------

  moveLeft(): boolean {
    return this.tryTranslate(-1, 0)
  }

  moveRight(): boolean {
    return this.tryTranslate(+1, 0)
  }

  rotateCW(): boolean {
    return this.tryRotate('cw')
  }

  rotateCCW(): boolean {
    return this.tryRotate('ccw')
  }

  softDrop(active: boolean): void {
    if (this.softDropActive === active) return
    this.softDropActive = active
    // Restart the gravity bucket so toggling between rates doesn't carry
    // accumulated time at the previous interval. Negligible effect at
    // human time scales but keeps semantics tidy.
    this.gravityAccumMs = 0
  }

  hardDrop(): boolean {
    if (this.phase !== EnginePhase.Falling && this.phase !== EnginePhase.Lock) {
      return false
    }
    if (!this.currentPiece) return false

    let piece = this.currentPiece
    let dropped = 0
    while (true) {
      const candidate = translate(piece, 0, -1)
      if (this.matrix.collides(candidate)) break
      piece = candidate
      dropped += 1
    }
    this.currentPiece = piece
    this.emit({ type: 'piece-hard-dropped', piece, cellsFallen: dropped })
    this.commitLock()
    return true
  }

  /**
   * Hold queue (§2.5). Swaps the active piece with the hold slot, or fills hold
   * and pulls the next bag piece. At most once per lock cycle.
   */
  hold(): boolean {
    if (this.phase !== EnginePhase.Falling && this.phase !== EnginePhase.Lock) {
      return false
    }
    if (!this.currentPiece || this.holdUsedForCurrentPiece) return false

    if (this.holdSlot === null) {
      const peeked = this.bag.peek(1)[0]
      if (peeked === undefined) return false
      const fresh = spawn(peeked)
      if (this.matrix.collides(fresh)) {
        this.setGameOver('blockOut')
        return false
      }
      this.holdSlot = this.currentPiece.type
      this.bag.next()
      if (!this.activateSpawnedPiece(fresh)) return false
    } else {
      const fromHold = this.holdSlot
      const fresh = spawn(fromHold)
      if (this.matrix.collides(fresh)) {
        this.setGameOver('blockOut')
        return false
      }
      this.holdSlot = this.currentPiece.type
      if (!this.activateSpawnedPiece(fresh)) return false
    }

    this.holdUsedForCurrentPiece = true
    this.emit({
      type: 'piece-held',
      piece: this.currentPiece,
      holdSlot: this.holdSlot,
    })
    return true
  }

  // -------------------------------------------------------------------------
  // Internal: phase transitions
  // -------------------------------------------------------------------------

  private spawnNext(): void {
    const pieceType = this.bag.next()
    const fresh = spawn(pieceType)
    this.activateSpawnedPiece(fresh)
  }

  /**
   * Places `fresh` (already typed spawn position) into play: block-out check,
   * `piece-generated`, optional §3.4 one-row drop, lock-down + phase.
   * @returns `false` if game over was triggered.
   */
  private activateSpawnedPiece(fresh: Tetrimino): boolean {
    if (this.matrix.collides(fresh)) {
      this.setGameOver('blockOut')
      return false
    }

    this.emit({ type: 'piece-generated', piece: fresh })

    let piece = fresh
    const droppedOnce = translate(piece, 0, -1)
    if (!this.matrix.collides(droppedOnce)) {
      piece = droppedOnce
    }

    this.currentPiece = piece
    this.lockDown.reset(bottomY(piece))
    this.gravityAccumMs = 0
    this.softDropActive = false

    if (canFall(this.matrix, piece)) {
      this.phase = EnginePhase.Falling
    } else {
      this.lockDown.onLanded()
      this.phase = EnginePhase.Lock
    }
    return true
  }

  private applyGravityStep(): void {
    if (!this.currentPiece) return
    const candidate = translate(this.currentPiece, 0, -1)
    if (this.matrix.collides(candidate)) {
      // Reached a Surface. Transition to Lock; the lock-down timer is
      // armed via `onLanded()`.
      this.lockDown.onLanded()
      this.phase = EnginePhase.Lock
      return
    }
    this.currentPiece = candidate
    this.lockDown.onFell(bottomY(this.currentPiece))
    if (this.softDropActive) {
      this.emit({ type: 'piece-soft-dropped', piece: this.currentPiece })
    }
    // Eagerly detect landing on the very same tick the piece reaches a
    // Surface, so state snapshots always reflect Lock the moment the piece
    // physically can't fall further (rather than waiting for the next
    // gravity step to fail).
    if (!canFall(this.matrix, this.currentPiece)) {
      this.lockDown.onLanded()
      this.phase = EnginePhase.Lock
    }
  }

  private tryTranslate(dx: number, dy: number): boolean {
    if (this.phase !== EnginePhase.Falling && this.phase !== EnginePhase.Lock) {
      return false
    }
    if (!this.currentPiece) return false

    const candidate = translate(this.currentPiece, dx, dy)
    if (this.matrix.collides(candidate)) return false

    this.currentPiece = candidate
    this.emit({ type: 'piece-moved', piece: this.currentPiece })
    this.lockDown.onMovedOrRotated()
    this.recheckSurface()
    return true
  }

  private tryRotate(direction: RotationDirection): boolean {
    if (this.phase !== EnginePhase.Falling && this.phase !== EnginePhase.Lock) {
      return false
    }
    if (!this.currentPiece) return false

    const result = srsTryRotate(this.matrix, this.currentPiece, direction)
    if (!result) return false

    this.currentPiece = result.piece
    this.emit({
      type: 'piece-rotated',
      piece: this.currentPiece,
      kickIndex: result.kickIndex,
    })
    this.lockDown.onMovedOrRotated()
    this.recheckSurface()
    return true
  }

  /**
   * After a successful move/rotate, decide whether the piece should switch
   * between Falling and Lock based on whether it can still fall. Crucial
   * for the §5.7 rules: a kick that lifts the piece off the ground pauses
   * lock-down; sliding back onto a ledge re-arms it.
   */
  private recheckSurface(): void {
    if (!this.currentPiece) return
    const stillFalls = canFall(this.matrix, this.currentPiece)
    if (this.phase === EnginePhase.Falling && !stillFalls) {
      this.lockDown.onLanded()
      this.phase = EnginePhase.Lock
    } else if (this.phase === EnginePhase.Lock && stillFalls) {
      this.lockDown.onLifted()
      this.phase = EnginePhase.Falling
      this.gravityAccumMs = 0
    }
  }

  /**
   * The synchronous Pattern → Eliminate → Completion sequence. Runs when
   * the lock-down timer expires or on hard drop.
   */
  private commitLock(): void {
    if (!this.currentPiece) return
    const lockedPiece = this.currentPiece

    this.matrix.lockPiece(lockedPiece)
    const cells = getOccupiedCells(lockedPiece)
    this.emit({ type: 'piece-locked', piece: lockedPiece, cells })

    const fullRows = this.matrix.findFullRows()
    if (fullRows.length > 0) {
      this.matrix.clearRows(fullRows)
      this.emit({ type: 'lines-cleared', rows: fullRows })
    }

    this.lines += fullRows.length
    while (this.lines >= this.goal && this.level < MAX_LEVEL) {
      this.level += 1
      this.goal += LINES_PER_LEVEL
      this.emit({ type: 'level-up', level: this.level })
    }

    this.currentPiece = null
    this.phase = EnginePhase.Generation
    this.generationTimerMs = this.generationDelayMs
    this.gravityAccumMs = 0
    this.softDropActive = false
    this.holdUsedForCurrentPiece = false
  }

  private setGameOver(reason: GameOverReason): void {
    this.phase = EnginePhase.GameOver
    this.gameOverReason = reason
    this.currentPiece = null
    this.emit({ type: 'game-over', reason })
  }

  private fallInterval(): number {
    return this.softDropActive ? softDropMsPerCell(this.level) : msPerCell(this.level)
  }

  private emit(event: EngineEvent): void {
    this.events.push(event)
  }

  private computeCanHold(): boolean {
    if (this.phase === EnginePhase.GameOver) return false
    if (this.phase !== EnginePhase.Falling && this.phase !== EnginePhase.Lock) return false
    if (!this.currentPiece) return false
    return !this.holdUsedForCurrentPiece
  }
}

// ---------------------------------------------------------------------------
// Module-local helpers
// ---------------------------------------------------------------------------

function bottomY(piece: Tetrimino): number {
  const cells = getOccupiedCells(piece)
  let min = Number.POSITIVE_INFINITY
  for (const c of cells) if (c.y < min) min = c.y
  return min
}

function canFall(matrix: Matrix, piece: Tetrimino): boolean {
  return !matrix.collides(translate(piece, 0, -1))
}

function clampLevel(level: number): number {
  return Math.max(1, Math.min(MAX_LEVEL, Math.floor(level)))
}
