/**
 * Shared types and constants for the Tetris engine.
 *
 * Coordinate system: bottom-left origin, 1-indexed (per 2009 Tetris Design
 * Guideline §A1.2.1). Cell (1,1) is the bottom-left of the visible Matrix,
 * (10,20) is the top-right of the visible area, and rows 21..40 are the
 * invisible Buffer Zone above the Skyline used for piece generation and
 * Game Over detection (§10).
 */

/**
 * Numeric value stored in a Matrix cell. `Empty` is 0 (the zero-initialized
 * Uint8Array default), and 1..7 are the seven Tetriminos in the order they
 * appear in §3.1 / §3.4 (the standard Guideline ordering used throughout this
 * codebase).
 */
export enum MinoType {
  Empty = 0,
  I = 1,
  O = 2,
  T = 3,
  S = 4,
  Z = 5,
  J = 6,
  L = 7,
}

/** A playable piece type — anything in `MinoType` except `Empty`. */
export type PieceType = Exclude<MinoType, MinoType.Empty>

/**
 * Rotation state (§A1.3). Rotate clockwise is `(f + 1) & 3`, counter-clockwise
 * is `(f + 3) & 3`. Pieces are always generated in the `N` facing (§3.4).
 */
export enum Facing {
  N = 0,
  E = 1,
  S = 2,
  W = 3,
}

/** A cell coordinate in the Matrix using guideline coords (1-indexed, bottom-left origin). */
export type Position = { x: number; y: number }

/** A byte stored in the Matrix grid. Always one of the `MinoType` values (0..7). */
export type Cell = number

/**
 * A live Tetrimino (the falling piece). Immutable value type — all engine
 * helpers (`rotate`, `translate`) return a new object rather than mutating.
 *
 * `origin` is the piece's Rotation Point 1 (visual rotation center) per §A1.4.
 * The four occupied cells are derived by applying `SHAPE_OFFSETS[type][facing]`
 * to `origin` (see `Tetrimino.ts`).
 */
export type Tetrimino = {
  type: PieceType
  facing: Facing
  origin: Position
}

export const MATRIX_WIDTH = 10
export const MATRIX_VISIBLE_HEIGHT = 20
export const MATRIX_BUFFER_HEIGHT = 20
export const MATRIX_TOTAL_HEIGHT = MATRIX_VISIBLE_HEIGHT + MATRIX_BUFFER_HEIGHT

/**
 * The Skyline (§A1.2.1). The top row of the visible Matrix is `y = SKYLINE_Y`;
 * any cell with `y > SKYLINE_Y` is in the Buffer Zone.
 */
export const SKYLINE_Y = MATRIX_VISIBLE_HEIGHT

/**
 * Standard Tetris Guideline colors for each piece (§3.1).
 * Defined here next to `MinoType` so the renderer and the engine share a single
 * source of truth.
 */
export const MINO_COLORS: Readonly<Record<PieceType, string>> = {
  [MinoType.I]: '#00F0F0', // light blue
  [MinoType.O]: '#F0F000', // yellow
  [MinoType.T]: '#A000F0', // purple
  [MinoType.S]: '#00F000', // green
  [MinoType.Z]: '#F00000', // red
  [MinoType.J]: '#0000F0', // dark blue
  [MinoType.L]: '#F0A000', // orange
}

// ---------------------------------------------------------------------------
// Engine state machine (slice 3)
// ---------------------------------------------------------------------------

/**
 * The time-bearing phases of the engine loop. The Guideline §A1.1 lists eight
 * phases (Generation, Falling, Lock, Pattern, Iterate, Animate, Eliminate,
 * Completion); all but Generation, Falling, and Lock resolve synchronously
 * inside a single `update()` call, so the engine only exposes these four:
 *
 * - `Generation` — waiting `GENERATION_DELAY_MS` after the previous lock-down
 *   before the next piece spawns (§A1.2.1).
 * - `Falling`    — a piece is in play and able to fall.
 * - `Lock`       — piece is on a Surface and the lock-down timer is counting.
 * - `GameOver`   — terminal; `update()` is a no-op.
 */
export enum EnginePhase {
  Generation = 'generation',
  Falling = 'falling',
  Lock = 'lock',
  GameOver = 'gameOver',
}

/**
 * Reason why the engine entered `GameOver`. Slice 3 implements `'blockOut'`
 * only (§10 a, spawn collision). `'lockOut'` and `'topOut'` are wired in
 * slice 8 alongside variants.
 */
export type GameOverReason = 'blockOut' | 'lockOut' | 'topOut'

/** Configurable parameters for an `Engine` instance. All fields optional. */
export type EngineConfig = {
  /** Seed for the 7-bag PRNG. Defaults to a time-based seed (non-deterministic). */
  seed?: number
  /** Starting level (clamped to 1..MAX_LEVEL). Defaults to 1. */
  startLevel?: number
  /** Lock-down delay in ms. Defaults to `LOCK_DELAY_MS`. */
  lockDelayMs?: number
  /** Max lock-down resets per piece (Extended Placement cap). Defaults to `MAX_LOCK_RESETS`. */
  maxLockResets?: number
  /** Delay between a piece locking and the next spawn. Defaults to `GENERATION_DELAY_MS`. */
  generationDelayMs?: number
  /** Length of the Next Queue look-ahead exposed in state snapshots. Defaults to 5. */
  nextQueueSize?: number
}

/** Public snapshot of the lock-down policy, suitable for rendering / debugging. */
export type LockDownSnapshot = {
  active: boolean
  timerMs: number
  movesLeft: number
  lowestBottomY: number
}

/**
 * Read-only snapshot of the engine state, suitable for passing to a renderer
 * or asserting on in tests. Always represents the engine after the most recent
 * `update()` / input call.
 */
export type EngineState = {
  phase: EnginePhase
  currentPiece: Tetrimino | null
  nextQueue: readonly PieceType[]
  /** Piece in the Hold slot, or `null` if empty. */
  holdPiece: PieceType | null
  /** `true` when the player may use Hold for the current falling piece (§2.5). */
  canHold: boolean
  level: number
  lines: number
  /** Total lines required to reach the *next* level (cumulative). */
  goal: number
  softDropActive: boolean
  /** Remaining ms before the next piece spawns (only meaningful in `Generation`). */
  generationTimerMs: number
  lockDown: LockDownSnapshot
  gameOver: boolean
  gameOverReason?: GameOverReason
}

/**
 * Discriminated union of events emitted by the engine since the last
 * `drainEvents()`. Consumers (renderer, audio, score) attach to specific
 * `type` variants; T-Spin / scoring tags will be added in slice 7.
 */
export type EngineEvent =
  | { type: 'piece-generated'; piece: Tetrimino }
  | { type: 'piece-moved'; piece: Tetrimino }
  | { type: 'piece-rotated'; piece: Tetrimino; kickIndex: number }
  | { type: 'piece-soft-dropped'; piece: Tetrimino }
  | { type: 'piece-hard-dropped'; piece: Tetrimino; cellsFallen: number }
  | { type: 'piece-held'; piece: Tetrimino; holdSlot: PieceType | null }
  | { type: 'piece-locked'; piece: Tetrimino; cells: readonly Position[] }
  | { type: 'lines-cleared'; rows: readonly number[] }
  | { type: 'level-up'; level: number }
  | { type: 'game-over'; reason: GameOverReason }

/** Lock-down delay in milliseconds (§5.7). */
export const LOCK_DELAY_MS = 500

/** Extended Placement: max successful moves/rotations that reset the timer per piece (§5.7). */
export const MAX_LOCK_RESETS = 15

/** Delay between a piece locking and the next piece spawning (§A1.2.1). */
export const GENERATION_DELAY_MS = 200

/** Soft-drop multiplier vs. natural gravity (§5.5). */
export const SOFT_DROP_MULTIPLIER = 20

/** Lines required per level in the Fixed Goal System (§6). */
export const LINES_PER_LEVEL = 10

/** Top of the Fixed Goal level scale (§6 / §7). Gravity is clamped at this level. */
export const MAX_LEVEL = 15

// ---------------------------------------------------------------------------
// Input layer (slice 4)
// ---------------------------------------------------------------------------

/**
 * Logical player commands. The `InputController` (DAS/ARR + state machine)
 * speaks this vocabulary; a `KeyboardAdapter` translates raw `KeyboardEvent`
 * codes into these. Pause is handled in the session layer (slice 6), not here.
 */
export enum InputCommand {
  MoveLeft = 'moveLeft',
  MoveRight = 'moveRight',
  SoftDrop = 'softDrop',
  HardDrop = 'hardDrop',
  RotateCW = 'rotateCW',
  RotateCCW = 'rotateCCW',
  Hold = 'hold',
}

/** Tunable timings for the `InputController`. All fields optional. */
export type InputConfig = {
  /** Delayed Auto Shift: ms held before auto-repeat begins. Default `DEFAULT_DAS_MS`. */
  dasMs?: number
  /**
   * Auto-Repeat Rate: ms between auto-repeated shifts after DAS expires.
   * `0` means "fire moves in a single burst until a wall is hit" (capped by
   * `MAX_ARR_BURST`). Default `DEFAULT_ARR_MS`.
   */
  arrMs?: number
  /**
   * When `true`, `press` and `update` are no-ops (e.g. game paused). `release`,
   * `releaseAll`, and externally invoked cleanup are not blocked.
   */
  isInputBlocked?: () => boolean
}

/**
 * Mapping from `KeyboardEvent.code` strings to logical `InputCommand` values.
 * Used by the `KeyboardAdapter` and freely replaceable per-instance (e.g. for
 * user-defined keymaps later).
 */
export type KeyMap = Readonly<Record<string, InputCommand>>

/** Default Delayed Auto Shift: 10 frames at 60 fps. */
export const DEFAULT_DAS_MS = 167

/** Default Auto-Repeat Rate: 2 frames at 60 fps. */
export const DEFAULT_ARR_MS = 33

/**
 * Safety cap for the `ARR = 0` burst path. With `ARR = 0`, a single
 * `InputController.update()` would otherwise loop until the move fails; this
 * upper bound stops runaway loops if the engine's move/collision logic ever
 * lies about a successful move.
 */
export const MAX_ARR_BURST = 50
