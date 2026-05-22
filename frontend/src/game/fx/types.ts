import type { PieceType } from '@/game/types'

/** One locked cell on a row about to clear (pre-`clearRows` snapshot). */
export type ClearedCellSnapshot = {
  x: number
  y: number
  color: PieceType
}

export type NotificationTone = 'default' | 'tetris' | 'tspin' | 'backToBack' | 'level'

export type ActionNotification = {
  id: string
  text: string
  tone: NotificationTone
  expiresAt: number
}

export type LineClearFx = {
  clearedCells: readonly ClearedCellSnapshot[]
  expiresAt: number
}

export type HardDropTrailFx = {
  minX: number
  maxX: number
  fromY: number
  toY: number
  color: string
  expiresAt: number
}

export const LINE_CLEAR_FX_MS = 350
export const HARD_DROP_TRAIL_MS = 150
export const NOTIFICATION_TTL_MS = 1200
