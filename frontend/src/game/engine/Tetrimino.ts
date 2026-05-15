/**
 * Tetrimino geometry and value-level helpers.
 *
 * All piece shapes are stored as precomputed `(dx, dy)` offset tables relative
 * to the piece's Rotation Point 1 (the visual rotation center described in
 * §A1.4). This lets us transcribe the §A1.3 facings directly and verify them
 * against the Guideline with simple fixture tests; it also keeps rotation a
 * pure table lookup at runtime (the SRS wall-kick logic in a later slice will
 * sit on top of this, not replace it).
 *
 * Offsets are listed in canonical order: y ascending, then x ascending within
 * each row (so the bottom-left mino comes first, the top-right comes last).
 *
 * I-tetrimino note: the I-piece's true SRS rotation center sits between cells
 * (it has a 4x4 bounding box). We use an integer origin on the second-from-left
 * mino of the N facing, and the offset tables describe each facing's position
 * inside that 4x4 bounding box — i.e. these are the canonical SRS positions,
 * not a "rotate around integer origin" approximation. This is what lets us
 * reuse the standard public SRS kick tables verbatim in slice 2.
 */

import {
  type Facing as FacingType,
  Facing,
  MinoType,
  type PieceType,
  type Position,
  type Tetrimino,
} from '@/game/types'

/**
 * A `(dx, dy)` offset from a piece's origin to one of its four minos.
 * Tuple form (not `Position`) so we don't allocate four objects per lookup.
 */
export type Offset = readonly [number, number]

/** Four offsets describing all minos of a piece in a given facing. */
export type ShapeOffsets = readonly [Offset, Offset, Offset, Offset]

/**
 * Per-piece, per-facing offset tables. Derived by hand from §A1.3 (the
 * Tetrimino Facings figure), with order canonicalized to (y asc, then x asc).
 */
export const SHAPE_OFFSETS: Readonly<Record<PieceType, Readonly<Record<FacingType, ShapeOffsets>>>> =
  {
    // I-tetrimino: integer origin on the second-from-left mino of the N facing.
    // N occupies cells (origin.x - 1 .. origin.x + 2, origin.y).
    [MinoType.I]: {
      [Facing.N]: [
        [-1, 0],
        [0, 0],
        [1, 0],
        [2, 0],
      ],
      [Facing.E]: [
        [1, -2],
        [1, -1],
        [1, 0],
        [1, 1],
      ],
      [Facing.S]: [
        [-1, -1],
        [0, -1],
        [1, -1],
        [2, -1],
      ],
      [Facing.W]: [
        [0, -2],
        [0, -1],
        [0, 0],
        [0, 1],
      ],
    },

    // O-tetrimino: §A1.4 specifies "all rotation points for all start positions
    // are the same" — the O never visually rotates. Origin is the bottom-left
    // mino so the 2x2 sits at (x..x+1, y..y+1).
    [MinoType.O]: {
      [Facing.N]: [
        [0, 0],
        [1, 0],
        [0, 1],
        [1, 1],
      ],
      [Facing.E]: [
        [0, 0],
        [1, 0],
        [0, 1],
        [1, 1],
      ],
      [Facing.S]: [
        [0, 0],
        [1, 0],
        [0, 1],
        [1, 1],
      ],
      [Facing.W]: [
        [0, 0],
        [1, 0],
        [0, 1],
        [1, 1],
      ],
    },

    // T-tetrimino: three across the row, one above the center.
    [MinoType.T]: {
      [Facing.N]: [
        [-1, 0],
        [0, 0],
        [1, 0],
        [0, 1],
      ],
      [Facing.E]: [
        [0, -1],
        [0, 0],
        [1, 0],
        [0, 1],
      ],
      [Facing.S]: [
        [0, -1],
        [-1, 0],
        [0, 0],
        [1, 0],
      ],
      [Facing.W]: [
        [0, -1],
        [-1, 0],
        [0, 0],
        [0, 1],
      ],
    },

    // S-tetrimino: two stacked horizontal dominoes, top offset right.
    [MinoType.S]: {
      [Facing.N]: [
        [-1, 0],
        [0, 0],
        [0, 1],
        [1, 1],
      ],
      [Facing.E]: [
        [1, -1],
        [0, 0],
        [1, 0],
        [0, 1],
      ],
      [Facing.S]: [
        [-1, -1],
        [0, -1],
        [0, 0],
        [1, 0],
      ],
      [Facing.W]: [
        [0, -1],
        [-1, 0],
        [0, 0],
        [-1, 1],
      ],
    },

    // Z-tetrimino: two stacked horizontal dominoes, top offset left.
    [MinoType.Z]: {
      [Facing.N]: [
        [0, 0],
        [1, 0],
        [-1, 1],
        [0, 1],
      ],
      [Facing.E]: [
        [0, -1],
        [0, 0],
        [1, 0],
        [1, 1],
      ],
      [Facing.S]: [
        [0, -1],
        [1, -1],
        [-1, 0],
        [0, 0],
      ],
      [Facing.W]: [
        [-1, -1],
        [-1, 0],
        [0, 0],
        [0, 1],
      ],
    },

    // J-tetrimino: hook on the upper-left in N.
    [MinoType.J]: {
      [Facing.N]: [
        [-1, 0],
        [0, 0],
        [1, 0],
        [-1, 1],
      ],
      [Facing.E]: [
        [0, -1],
        [0, 0],
        [0, 1],
        [1, 1],
      ],
      [Facing.S]: [
        [1, -1],
        [-1, 0],
        [0, 0],
        [1, 0],
      ],
      [Facing.W]: [
        [-1, -1],
        [0, -1],
        [0, 0],
        [0, 1],
      ],
    },

    // L-tetrimino: hook on the upper-right in N.
    [MinoType.L]: {
      [Facing.N]: [
        [-1, 0],
        [0, 0],
        [1, 0],
        [1, 1],
      ],
      [Facing.E]: [
        [0, -1],
        [1, -1],
        [0, 0],
        [0, 1],
      ],
      [Facing.S]: [
        [-1, -1],
        [-1, 0],
        [0, 0],
        [1, 0],
      ],
      [Facing.W]: [
        [0, -1],
        [0, 0],
        [-1, 1],
        [0, 1],
      ],
    },
  }

/**
 * Spawn origin per §3.4 / §A1.2.1. Pieces always generate north-facing on the
 * lower of the two spawn rows (y = 21, the first row of the Buffer Zone). The
 * Guideline says "the tetrimino drops one row if no existing Block is in its
 * path" immediately after generation — that's the engine's job, not ours.
 *
 * - T, S, Z, J, L (3-wide): origin at column 5 so minos cover columns 4..6.
 * - I (4-wide): origin at column 5 so minos cover columns 4..7.
 * - O (2-wide): origin at column 5 so minos cover columns 5..6.
 */
export const SPAWN_ORIGIN: Readonly<Record<PieceType, Position>> = {
  [MinoType.I]: { x: 5, y: 21 },
  [MinoType.O]: { x: 5, y: 21 },
  [MinoType.T]: { x: 5, y: 21 },
  [MinoType.S]: { x: 5, y: 21 },
  [MinoType.Z]: { x: 5, y: 21 },
  [MinoType.J]: { x: 5, y: 21 },
  [MinoType.L]: { x: 5, y: 21 },
}

/** Lookup helper — `noUncheckedIndexedAccess` makes Record lookups `T | undefined`, so we centralize the assertion. */
export function getMinoOffsets(type: PieceType, facing: FacingType): ShapeOffsets {
  const perFacing = SHAPE_OFFSETS[type]
  const offsets = perFacing[facing]
  return offsets
}

/**
 * Returns the four matrix cells the piece currently occupies, in canonical
 * (y asc, x asc) order. Always exactly four elements.
 */
export function getOccupiedCells(piece: Tetrimino): Position[] {
  const offsets = getMinoOffsets(piece.type, piece.facing)
  const { x, y } = piece.origin
  return offsets.map(([dx, dy]) => ({ x: x + dx, y: y + dy }))
}

/** Returns a fresh north-facing piece at its spawn position (§3.4). */
export function spawn(type: PieceType): Tetrimino {
  return {
    type,
    facing: Facing.N,
    origin: { ...SPAWN_ORIGIN[type] },
  }
}

/**
 * Rotates the piece 90 degrees. Returns a new piece with the new `facing`;
 * `origin` is unchanged. SRS wall-kicks are intentionally NOT applied here —
 * they live in `SRS.ts` (slice 2) and operate on top of this function.
 */
export function rotate(piece: Tetrimino, direction: 'cw' | 'ccw'): Tetrimino {
  const delta = direction === 'cw' ? 1 : 3
  return {
    type: piece.type,
    facing: ((piece.facing + delta) & 3) as FacingType,
    origin: { ...piece.origin },
  }
}

/** Returns a new piece with the origin shifted by `(dx, dy)`. */
export function translate(piece: Tetrimino, dx: number, dy: number): Tetrimino {
  return {
    type: piece.type,
    facing: piece.facing,
    origin: { x: piece.origin.x + dx, y: piece.origin.y + dy },
  }
}
