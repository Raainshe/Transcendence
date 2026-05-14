/**
 * Ghost piece — where the active tetrimino would rest if dropped straight down.
 * Pure geometry on top of `Matrix.collides` + `translate`; no engine coupling.
 */

import type { Matrix } from '@/game/engine/Matrix'
import { translate } from '@/game/engine/Tetrimino'
import type { Tetrimino } from '@/game/types'

/** Lowest resting position of `piece` with the same facing (SRS ghost). */
export function computeGhostPiece(matrix: Matrix, piece: Tetrimino): Tetrimino {
  let ghost = piece
  while (!matrix.collides(translate(ghost, 0, -1))) {
    ghost = translate(ghost, 0, -1)
  }
  return ghost
}
