import { cellAt } from '@/game/engine/matrixSnapshot'
import {
  MATRIX_VISIBLE_HEIGHT,
  MATRIX_WIDTH,
  MINO_COLORS,
  MinoType,
  type PieceType,
} from '@/game/types'

function colorForCell(value: number): string | null {
  if (value === MinoType.Empty) return null
  return MINO_COLORS[value as PieceType]
}

function rowFromTop(matrixY: number): number {
  return MATRIX_VISIBLE_HEIGHT - matrixY
}

export function drawLockedMatrix(
  ctx: CanvasRenderingContext2D,
  grid: Uint8Array,
  cellPx: number,
  visibleOnly = true,
): void {
  const topY = visibleOnly ? MATRIX_VISIBLE_HEIGHT : 40
  const w = MATRIX_WIDTH * cellPx
  const h = (visibleOnly ? MATRIX_VISIBLE_HEIGHT : 40) * cellPx

  ctx.clearRect(0, 0, w, h)
  ctx.fillStyle = 'rgba(0, 0, 0, 0.45)'
  ctx.fillRect(0, 0, w, h)

  ctx.strokeStyle = 'rgba(255, 255, 255, 0.06)'
  ctx.lineWidth = 1
  for (let gx = 0; gx <= MATRIX_WIDTH; gx++) {
    ctx.beginPath()
    ctx.moveTo(gx * cellPx, 0)
    ctx.lineTo(gx * cellPx, h)
    ctx.stroke()
  }
  for (let gy = 0; gy <= (visibleOnly ? MATRIX_VISIBLE_HEIGHT : 40); gy++) {
    ctx.beginPath()
    ctx.moveTo(0, gy * cellPx)
    ctx.lineTo(w, gy * cellPx)
    ctx.stroke()
  }

  const pad = 1
  for (let y = 1; y <= topY; y++) {
    for (let x = 1; x <= MATRIX_WIDTH; x++) {
      const v = cellAt(grid, x, y)
      const c = colorForCell(v)
      if (!c) continue
      ctx.fillStyle = c
      ctx.fillRect(
        (x - 1) * cellPx + pad,
        rowFromTop(y) * cellPx + pad,
        cellPx - pad * 2,
        cellPx - pad * 2,
      )
    }
  }
}
