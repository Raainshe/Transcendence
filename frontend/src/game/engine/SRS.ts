/**
 * Super Rotation System — kick tables + `tryRotate`.
 *
 * The Guideline (§A1.4) defines five "Rotation Points" per quarter-turn. SRS
 * tries them in order; the first that doesn't collide is the result of the
 * rotation. Different kick indices correspond to the Guideline's "Visual",
 * "Off the wall", "Off the floor", and "Out of a well" cases (§A1.4 a-f).
 *
 * Coordinate convention: bottom-left origin, y+ up (matches the rest of the
 * engine). The canonical public SRS tables on Tetris Wiki / Hard Drop wiki
 * are written with y+ DOWN; the offsets stored here are those tables with the
 * y components negated. Cross-reference:
 *   https://tetris.wiki/Super_Rotation_System
 *
 * The kick index reported by `tryRotate` (0..4) directly corresponds to
 * Rotation Point 1..5 in the Guideline. Slice 7 (T-Spin detection) uses this:
 * §9.1.1 says "Rotation Point 5 is used to rotate the tetrimino into the
 * T-Slot" always counts as a full T-Spin, not a Mini T-Spin.
 */

import { type Offset, rotate, translate } from '@/game/engine/Tetrimino'
import { Facing, MinoType, type PieceType, type Tetrimino } from '@/game/types'

import type { Matrix } from '@/game/engine/Matrix'

/** Five candidate offsets — Guideline "Rotation Points 1..5". */
type KickList = readonly [Offset, Offset, Offset, Offset, Offset]

/**
 * 4x4 grid indexed by `[from][to]`. Only the 8 quarter-turn pairs are
 * populated (N<->E, E<->S, S<->W, W<->N); the other 8 are `undefined`
 * because SRS does not define half-turns or self-rotations.
 */
type KickTable = readonly [
  readonly (KickList | undefined)[],
  readonly (KickList | undefined)[],
  readonly (KickList | undefined)[],
  readonly (KickList | undefined)[],
]

/**
 * JLSTZ pieces (the 3x3 bounding-box pieces) share one table. Values mirror
 * the canonical SRS table with the y axis negated to match our y+up convention.
 */
const JLSTZ_KICKS: KickTable = [
  // from N (0)
  [
    undefined,
    [
      [0, 0],
      [-1, 0],
      [-1, -1],
      [0, +2],
      [-1, +2],
    ], // N -> E
    undefined,
    [
      [0, 0],
      [+1, 0],
      [+1, -1],
      [0, +2],
      [+1, +2],
    ], // N -> W
  ],
  // from E (1)
  [
    [
      [0, 0],
      [+1, 0],
      [+1, +1],
      [0, -2],
      [+1, -2],
    ], // E -> N
    undefined,
    [
      [0, 0],
      [+1, 0],
      [+1, +1],
      [0, -2],
      [+1, -2],
    ], // E -> S
    undefined,
  ],
  // from S (2)
  [
    undefined,
    [
      [0, 0],
      [-1, 0],
      [-1, -1],
      [0, +2],
      [-1, +2],
    ], // S -> E
    undefined,
    [
      [0, 0],
      [+1, 0],
      [+1, -1],
      [0, +2],
      [+1, +2],
    ], // S -> W
  ],
  // from W (3)
  [
    [
      [0, 0],
      [-1, 0],
      [-1, +1],
      [0, -2],
      [-1, -2],
    ], // W -> N
    undefined,
    [
      [0, 0],
      [-1, 0],
      [-1, +1],
      [0, -2],
      [-1, -2],
    ], // W -> S
    undefined,
  ],
]

/**
 * I-tetrimino has its own kick table because of its 4x4 bounding box. Values
 * again y-negated from the public canonical table.
 */
const I_KICKS: KickTable = [
  // from N (0)
  [
    undefined,
    [
      [0, 0],
      [-2, 0],
      [+1, 0],
      [-2, +1],
      [+1, -2],
    ], // N -> E
    undefined,
    [
      [0, 0],
      [-1, 0],
      [+2, 0],
      [-1, -2],
      [+2, +1],
    ], // N -> W
  ],
  // from E (1)
  [
    [
      [0, 0],
      [+2, 0],
      [-1, 0],
      [+2, -1],
      [-1, +2],
    ], // E -> N
    undefined,
    [
      [0, 0],
      [-1, 0],
      [+2, 0],
      [-1, -2],
      [+2, +1],
    ], // E -> S
    undefined,
  ],
  // from S (2)
  [
    undefined,
    [
      [0, 0],
      [+1, 0],
      [-2, 0],
      [+1, +2],
      [-2, -1],
    ], // S -> E
    undefined,
    [
      [0, 0],
      [+2, 0],
      [-1, 0],
      [+2, -1],
      [-1, +2],
    ], // S -> W
  ],
  // from W (3)
  [
    [
      [0, 0],
      [+1, 0],
      [-2, 0],
      [+1, +2],
      [-2, -1],
    ], // W -> N
    undefined,
    [
      [0, 0],
      [-2, 0],
      [+1, 0],
      [-2, +1],
      [+1, -2],
    ], // W -> S
    undefined,
  ],
]

/**
 * O-piece kick list. §A1.4 specifies that "all rotation points for the O
 * Tetrimino are the same" — visually it doesn't rotate. We still let the
 * player "rotate" the O so the lock-down timer can reset normally (§5.7);
 * the kick is always `(0, 0)`.
 */
const O_KICKS: readonly Offset[] = [[0, 0]]

/** Direction of a 90-degree rotation, as the player perceives it. */
export type RotationDirection = 'cw' | 'ccw'

/** Result of a successful rotation. `kickIndex` is 0..4 = Guideline Point 1..5. */
export type RotationResult = {
  piece: Tetrimino
  kickIndex: number
}

/**
 * Returns the kick list for `(type, from -> to)`. Empty array for invalid
 * transitions (same facing or half-turn). The O-tetrimino always returns
 * a single `(0, 0)` kick regardless of `from`/`to`.
 */
export function getKicks(type: PieceType, from: Facing, to: Facing): readonly Offset[] {
  if (type === MinoType.O) return O_KICKS
  const table = type === MinoType.I ? I_KICKS : JLSTZ_KICKS
  const row = table[from]
  if (!row) return []
  return row[to] ?? []
}

/** Computes the target facing for a CW or CCW quarter-turn. */
export function getTargetFacing(from: Facing, direction: RotationDirection): Facing {
  const delta = direction === 'cw' ? 1 : 3
  return ((from + delta) & 3) as Facing
}

/**
 * Attempts a 90-degree rotation, trying each kick in order.
 *
 * Returns `{ piece, kickIndex }` for the first non-colliding candidate, or
 * `null` if all five (or one, for O) kicks collide. Pure: never mutates the
 * matrix or the input piece.
 */
export function tryRotate(
  matrix: Matrix,
  piece: Tetrimino,
  direction: RotationDirection,
): RotationResult | null {
  const targetFacing = getTargetFacing(piece.facing, direction)
  const kicks = getKicks(piece.type, piece.facing, targetFacing)
  if (kicks.length === 0) return null

  const rotated = rotate(piece, direction)
  for (let i = 0; i < kicks.length; i++) {
    const kick = kicks[i]
    if (!kick) continue
    const candidate = translate(rotated, kick[0], kick[1])
    if (!matrix.collides(candidate)) {
      return { piece: candidate, kickIndex: i }
    }
  }
  return null
}
