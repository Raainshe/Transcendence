/**
 * T-Spin detection at lock (§9): 3-corner rule, Mini vs Full, kick index 4 ⇒ Full.
 */

import type { Matrix } from '@/game/engine/Matrix'
import { Facing, MinoType, type Tetrimino } from '@/game/types'

import type { TSpinKind } from '@/game/scoring/types'

export type TSpinResult = {
  kind: TSpinKind
  cornersFilled: number
}

/** Four diagonal corners of the 3×3 T-slot around the rotation point. */
function cornerCells(originX: number, originY: number): readonly { x: number; y: number }[] {
  return [
    { x: originX - 1, y: originY - 1 },
    { x: originX + 1, y: originY - 1 },
    { x: originX - 1, y: originY + 1 },
    { x: originX + 1, y: originY + 1 },
  ]
}

/** Front two corners by facing (§9 — "front" of the T). */
function frontCornerIndices(facing: Facing): readonly [number, number] {
  switch (facing) {
    case Facing.N:
      return [2, 3]
    case Facing.E:
      return [1, 3]
    case Facing.S:
      return [0, 1]
    case Facing.W:
      return [0, 2]
    default:
      return [0, 1]
  }
}

function isCornerFilled(matrix: Matrix, x: number, y: number): boolean {
  if (x < 1 || x > 10 || y < 1) return true
  return !matrix.isCellEmpty(x, y)
}

/**
 * Detect T-Spin at lock. Matrix must already include the locked T piece.
 * `lastActionWasRotation` must be true for a valid T-Spin.
 */
export function detectTSpin(
  matrix: Matrix,
  piece: Tetrimino,
  lastActionWasRotation: boolean,
  lastKickIndex: number | null,
): TSpinResult {
  if (piece.type !== MinoType.T || !lastActionWasRotation) {
    return { kind: 'none', cornersFilled: 0 }
  }

  const { x: ox, y: oy } = piece.origin
  const corners = cornerCells(ox, oy)
  let filled = 0
  for (const c of corners) {
    if (isCornerFilled(matrix, c.x, c.y)) filled += 1
  }

  if (filled < 3) {
    return { kind: 'none', cornersFilled: filled }
  }

  if (lastKickIndex === 4) {
    return { kind: 'full', cornersFilled: filled }
  }

  const [i0, i1] = frontCornerIndices(piece.facing)
  const frontFilled =
    isCornerFilled(matrix, corners[i0]!.x, corners[i0]!.y) &&
    isCornerFilled(matrix, corners[i1]!.x, corners[i1]!.y)

  return {
    kind: frontFilled ? 'full' : 'mini',
    cornersFilled: filled,
  }
}
