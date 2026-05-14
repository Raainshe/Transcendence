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
