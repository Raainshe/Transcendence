/**
 * The Matrix — the rectangular play area where locked Blocks live (§A1.2.1).
 *
 * Storage: a single `Uint8Array` of `MATRIX_WIDTH * MATRIX_TOTAL_HEIGHT = 400`
 * bytes, row-major in guideline coords. Cell `(x, y)` (1-indexed, bottom-left
 * origin) is stored at index `(y - 1) * 10 + (x - 1)`. The byte value is one of
 * `MinoType` (0 = empty, 1..7 = locked piece type / color).
 *
 * The engine never reads this array directly; everything goes through `get()`
 * and `set()`. That lets us swap storage later (e.g. row-typed arrays for
 * faster line-clear shifts) without touching consumers.
 */

import { getOccupiedCells } from '@/game/engine/Tetrimino'
import {
  type Cell,
  MATRIX_TOTAL_HEIGHT,
  MATRIX_VISIBLE_HEIGHT,
  MATRIX_WIDTH,
  MinoType,
  type Tetrimino,
} from '@/game/types'

export class Matrix {
  private readonly cells: Uint8Array

  constructor() {
    this.cells = new Uint8Array(MATRIX_WIDTH * MATRIX_TOTAL_HEIGHT)
  }

  /**
   * Returns the cell value at `(x, y)` in guideline coords. For any
   * out-of-bounds position (walls, floor, or above the top of the Buffer Zone)
   * returns `MinoType.Empty`. Wall/floor-vs-above-buffer semantics for
   * collision detection live in `collides()`, not here, so this stays a pure
   * grid reader.
   */
  get(x: number, y: number): Cell {
    if (x < 1 || x > MATRIX_WIDTH || y < 1 || y > MATRIX_TOTAL_HEIGHT) {
      return MinoType.Empty
    }
    return this.cells[Matrix.idx(x, y)] ?? MinoType.Empty
  }

  /** Writes a cell. No-op if `(x, y)` is out of bounds. */
  set(x: number, y: number, value: Cell): void {
    if (x < 1 || x > MATRIX_WIDTH || y < 1 || y > MATRIX_TOTAL_HEIGHT) return
    this.cells[Matrix.idx(x, y)] = value
  }

  /** Convenience: `true` iff the cell at `(x, y)` is empty (or above the buffer). */
  isCellEmpty(x: number, y: number): boolean {
    return this.get(x, y) === MinoType.Empty
  }

  /**
   * True if any of `piece`'s four cells overlaps a wall (`x < 1` or
   * `x > MATRIX_WIDTH`), the floor (`y < 1`), or a locked Block in the Matrix.
   * Cells above the top of the Buffer Zone (`y > MATRIX_TOTAL_HEIGHT`) do NOT
   * count as collisions — pieces are allowed to exist above the Matrix per
   * §A1.2.3 / §10 (game-over is decided separately by the lock-down rules).
   */
  collides(piece: Tetrimino): boolean {
    const cells = getOccupiedCells(piece)
    for (const { x, y } of cells) {
      if (x < 1 || x > MATRIX_WIDTH) return true
      if (y < 1) return true
      if (y > MATRIX_TOTAL_HEIGHT) continue
      if (this.cells[Matrix.idx(x, y)] !== MinoType.Empty) return true
    }
    return false
  }

  /**
   * Writes the piece's four minos into the Matrix using `piece.type` as the
   * cell value. In dev builds this asserts that the target cells are empty
   * and inside the Matrix — calling `lockPiece` on a colliding piece is a bug.
   */
  lockPiece(piece: Tetrimino): void {
    const cells = getOccupiedCells(piece)
    if (import.meta.env?.DEV) {
      for (const { x, y } of cells) {
        if (x < 1 || x > MATRIX_WIDTH || y < 1 || y > MATRIX_TOTAL_HEIGHT) {
          throw new Error(`lockPiece: cell (${x}, ${y}) is out of bounds`)
        }
        if (this.cells[Matrix.idx(x, y)] !== MinoType.Empty) {
          throw new Error(`lockPiece: cell (${x}, ${y}) is already occupied`)
        }
      }
    }
    for (const { x, y } of cells) {
      if (x < 1 || x > MATRIX_WIDTH || y < 1 || y > MATRIX_TOTAL_HEIGHT) continue
      this.cells[Matrix.idx(x, y)] = piece.type
    }
  }

  /**
   * Returns the y-coordinates of every fully-filled row, ascending. Used by
   * the Pattern Phase (§A1.2.4) to feed `clearRows`. A row is "full" when all
   * `MATRIX_WIDTH` cells contain a non-zero `MinoType`.
   */
  findFullRows(): number[] {
    const full: number[] = []
    for (let y = 1; y <= MATRIX_TOTAL_HEIGHT; y++) {
      let isFull = true
      const rowStart = (y - 1) * MATRIX_WIDTH
      for (let x = 0; x < MATRIX_WIDTH; x++) {
        if (this.cells[rowStart + x] === MinoType.Empty) {
          isFull = false
          break
        }
      }
      if (isFull) full.push(y)
    }
    return full
  }

  /**
   * Removes the listed rows and gravity-collapses everything above (§A1.2.7).
   * Input rows are 1-indexed y-coordinates; duplicates and out-of-range values
   * are ignored. Topmost rows are zero-filled after the shift.
   */
  clearRows(rows: number[]): void {
    const drop = new Set<number>()
    for (const y of rows) {
      if (y >= 1 && y <= MATRIX_TOTAL_HEIGHT) drop.add(y)
    }
    if (drop.size === 0) return

    const kept: Uint8Array[] = []
    for (let y = 1; y <= MATRIX_TOTAL_HEIGHT; y++) {
      if (drop.has(y)) continue
      const rowStart = (y - 1) * MATRIX_WIDTH
      kept.push(this.cells.slice(rowStart, rowStart + MATRIX_WIDTH))
    }

    this.cells.fill(MinoType.Empty)
    for (let i = 0; i < kept.length; i++) {
      const row = kept[i]
      if (!row) continue
      this.cells.set(row, i * MATRIX_WIDTH)
    }
  }

  /** Deep copy. Useful for tests and any future rollback/replay logic. */
  clone(): Matrix {
    const copy = new Matrix()
    copy.cells.set(this.cells)
    return copy
  }

  /**
   * ASCII dump for debugging and test fixtures. Top of the rendered area is
   * the first line, bottom row (y = 1) is the last line. Empty cells are `.`;
   * filled cells use the first letter of the piece type (`I/O/T/S/Z/J/L`).
   */
  debugString(visibleOnly = true): string {
    const topY = visibleOnly ? MATRIX_VISIBLE_HEIGHT : MATRIX_TOTAL_HEIGHT
    const lines: string[] = []
    for (let y = topY; y >= 1; y--) {
      let line = ''
      for (let x = 1; x <= MATRIX_WIDTH; x++) {
        line += Matrix.glyph(this.get(x, y))
      }
      lines.push(line)
    }
    return lines.join('\n')
  }

  private static idx(x: number, y: number): number {
    return (y - 1) * MATRIX_WIDTH + (x - 1)
  }

  private static glyph(cell: Cell): string {
    switch (cell) {
      case MinoType.I:
        return 'I'
      case MinoType.O:
        return 'O'
      case MinoType.T:
        return 'T'
      case MinoType.S:
        return 'S'
      case MinoType.Z:
        return 'Z'
      case MinoType.J:
        return 'J'
      case MinoType.L:
        return 'L'
      default:
        return '.'
    }
  }
}
