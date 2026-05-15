/**
 * §10 game-over detection helpers.
 */

import { getOccupiedCells } from '@/game/engine/Tetrimino'
import type { Matrix } from '@/game/engine/Matrix'
import { MATRIX_WIDTH, MinoType, SKYLINE_Y, type Tetrimino } from '@/game/types'

/** Generation Area rows (buffer zone spawn rows per §A1.2.1). */
export const GENERATION_ROW_MIN = 21
export const GENERATION_ROW_MAX = 22

/** §10 (b) Lock-out: locked piece overlaps the Generation Area. */
export function isLockOut(lockedPiece: Tetrimino): boolean {
  for (const { x, y } of getOccupiedCells(lockedPiece)) {
    if (x >= 1 && x <= MATRIX_WIDTH && y >= GENERATION_ROW_MIN && y <= GENERATION_ROW_MAX) {
      return true
    }
  }
  return false
}

/** §10 (c) Top-out: any locked block above the Skyline (buffer zone). */
export function isTopOut(matrix: Matrix): boolean {
  for (let y = SKYLINE_Y + 1; y <= SKYLINE_Y + 20; y++) {
    for (let x = 1; x <= MATRIX_WIDTH; x++) {
      if (matrix.get(x, y) !== MinoType.Empty) return true
    }
  }
  return false
}
