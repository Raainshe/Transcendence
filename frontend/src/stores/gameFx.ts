import { defineStore } from 'pinia'
import { ref } from 'vue'

import {
  notificationFromLevelUp,
  notificationsFromScoreBreakdown,
} from '@/game/fx/notifications'
import type {
  ActionNotification,
  HardDropTrailFx,
  LineClearFx,
} from '@/game/fx/types'
import {
  HARD_DROP_TRAIL_MS,
  LINE_CLEAR_FX_MS,
} from '@/game/fx/types'
import { getOccupiedCells } from '@/game/engine/Tetrimino'
import type { EngineEvent } from '@/game/types'
import { MINO_COLORS } from '@/game/types'

let notificationCounter = 0

export const useGameFxStore = defineStore('gameFx', () => {
  const notifications = ref<ActionNotification[]>([])
  const lineClearFlash = ref<LineClearFx | null>(null)
  const hardDropTrail = ref<HardDropTrailFx | null>(null)

  function reset(): void {
    notifications.value = []
    lineClearFlash.value = null
    hardDropTrail.value = null
  }

  function prune(now: number): void {
    notifications.value = notifications.value.filter((n) => n.expiresAt > now)
    if (lineClearFlash.value && lineClearFlash.value.expiresAt <= now) {
      lineClearFlash.value = null
    }
    if (hardDropTrail.value && hardDropTrail.value.expiresAt <= now) {
      hardDropTrail.value = null
    }
  }

  function ingestEvents(events: readonly EngineEvent[]): void {
    const now = performance.now()

    for (const event of events) {
      switch (event.type) {
        case 'lines-cleared':
          if (event.clearedCells.length > 0) {
            lineClearFlash.value = {
              clearedCells: event.clearedCells,
              expiresAt: now + LINE_CLEAR_FX_MS,
            }
          }
          break
        case 'piece-hard-dropped': {
          const cells = getOccupiedCells(event.piece)
          const xs = cells.map((c) => c.x)
          if (xs.length === 0 || event.cellsFallen <= 0) break
          const fromY = event.piece.origin.y + event.cellsFallen
          const toY = event.piece.origin.y
          hardDropTrail.value = {
            minX: Math.min(...xs),
            maxX: Math.max(...xs),
            fromY,
            toY,
            color: MINO_COLORS[event.piece.type],
            expiresAt: now + HARD_DROP_TRAIL_MS,
          }
          break
        }
        case 'score-awarded': {
          const prefix = `n-${notificationCounter++}`
          const added = notificationsFromScoreBreakdown(event.breakdown, now, prefix)
          if (added.length > 0) {
            notifications.value = [...notifications.value, ...added].slice(-4)
          }
          break
        }
        case 'level-up':
          notifications.value = [
            ...notifications.value,
            notificationFromLevelUp(event.level, now, `lvl-${notificationCounter++}`),
          ].slice(-4)
          break
        default:
          break
      }
    }
  }

  return {
    notifications,
    lineClearFlash,
    hardDropTrail,
    ingestEvents,
    prune,
    reset,
  }
})
