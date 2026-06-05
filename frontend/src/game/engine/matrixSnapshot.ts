import type { Matrix } from '@/game/engine/Matrix'
import { MATRIX_WIDTH, MinoType } from '@/game/types'

export const MATRIX_CELL_COUNT = MATRIX_WIDTH * 40

function bytesToBase64(bytes: Uint8Array): string {
  let binary = ''
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]!)
  }
  return btoa(binary)
}

function base64ToBytes(encoded: string): Uint8Array | null {
  try {
    const binary = atob(encoded)
    if (binary.length !== MATRIX_CELL_COUNT) return null
    const out = new Uint8Array(MATRIX_CELL_COUNT)
    for (let i = 0; i < binary.length; i++) {
      const v = binary.charCodeAt(i)
      if (v > 7) return null
      out[i] = v
    }
    return out
  } catch {
    return null
  }
}

export function encodeLockedMatrixBase64(matrix: Matrix): string {
  return bytesToBase64(matrix.snapshotBytes())
}

export function decodeLockedMatrixBase64(board: string): Uint8Array | null {
  if (!board) return null
  return base64ToBytes(board)
}

/** Read a cell from a decoded snapshot grid (1-indexed guideline coords). */
export function cellAt(grid: Uint8Array, x: number, y: number): number {
  if (x < 1 || x > MATRIX_WIDTH || y < 1 || y > 40) return MinoType.Empty
  return grid[(y - 1) * MATRIX_WIDTH + (x - 1)] ?? MinoType.Empty
}
