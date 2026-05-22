import type { LineClearKind, ScoreBreakdown, TSpinKind } from '@/game/scoring/types'

import type { ActionNotification, NotificationTone } from '@/game/fx/types'
import { NOTIFICATION_TTL_MS } from '@/game/fx/types'

export type NotificationDraft = {
  text: string
  tone: NotificationTone
}

function lineClearLabel(kind: LineClearKind): NotificationDraft | null {
  switch (kind) {
    case 'tetris':
      return { text: 'TETRIS', tone: 'tetris' }
    case 'triple':
      return { text: 'TRIPLE', tone: 'default' }
    case 'double':
      return { text: 'DOUBLE', tone: 'default' }
    case 'single':
      return null
    default:
      return null
  }
}

function tSpinLabel(kind: TSpinKind, linesCleared: number): NotificationDraft | null {
  if (kind === 'none') return null
  if (kind === 'mini') {
    return { text: 'MINI T-SPIN', tone: 'tspin' }
  }
  if (linesCleared >= 3) return { text: 'T-SPIN TRIPLE', tone: 'tspin' }
  if (linesCleared === 2) return { text: 'T-SPIN DOUBLE', tone: 'tspin' }
  if (linesCleared === 1) return { text: 'T-SPIN', tone: 'tspin' }
  return { text: 'T-SPIN!', tone: 'tspin' }
}

/** Build HUD callouts from a scoring breakdown (line clear / T-spin / B2B). */
export function notificationsFromScoreBreakdown(
  breakdown: ScoreBreakdown,
  now: number,
  idPrefix: string,
): ActionNotification[] {
  const out: ActionNotification[] = []
  let seq = 0

  const push = (draft: NotificationDraft | null) => {
    if (!draft) return
    out.push({
      id: `${idPrefix}-${seq++}`,
      text: draft.text,
      tone: draft.tone,
      expiresAt: now + NOTIFICATION_TTL_MS,
    })
  }

  if (breakdown.reason === 'tSpinNoLines') {
    push({ text: 'T-SPIN!', tone: 'tspin' })
    return out
  }

  if (breakdown.reason !== 'lineClear') return out

  const tSpin = tSpinLabel(breakdown.tSpinKind, breakdown.linesCleared)
  if (tSpin) {
    push(tSpin)
  } else {
    push(lineClearLabel(breakdown.lineClearKind))
  }

  if (breakdown.backToBackMultiplier > 1 && breakdown.linesCleared > 0) {
    push({ text: 'BACK-TO-BACK', tone: 'backToBack' })
  }

  return out
}

export function notificationFromLevelUp(level: number, now: number, id: string): ActionNotification {
  return {
    id,
    text: `LEVEL ${level}`,
    tone: 'level',
    expiresAt: now + NOTIFICATION_TTL_MS,
  }
}
