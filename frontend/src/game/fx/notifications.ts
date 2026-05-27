import { i18n } from '@/i18n'
import type { LineClearKind, ScoreBreakdown, TSpinKind } from '@/game/scoring/types'

import type { ActionNotification, NotificationTone } from '@/game/fx/types'
import { NOTIFICATION_TTL_MS } from '@/game/fx/types'

export type NotificationDraft = {
  textKey: string
  textParams?: Record<string, number | string>
  tone: NotificationTone
}

function lineClearLabel(kind: LineClearKind): NotificationDraft | null {
  switch (kind) {
    case 'tetris':
      return { textKey: 'game.notifications.tetris', tone: 'tetris' }
    case 'triple':
      return { textKey: 'game.notifications.triple', tone: 'default' }
    case 'double':
      return { textKey: 'game.notifications.double', tone: 'default' }
    case 'single':
      return null
    default:
      return null
  }
}

function tSpinLabel(kind: TSpinKind, linesCleared: number): NotificationDraft | null {
  if (kind === 'none') return null
  if (kind === 'mini') {
    return { textKey: 'game.notifications.miniTSpin', tone: 'tspin' }
  }
  if (linesCleared >= 3) return { textKey: 'game.notifications.tSpinTriple', tone: 'tspin' }
  if (linesCleared === 2) return { textKey: 'game.notifications.tSpinDouble', tone: 'tspin' }
  if (linesCleared === 1) return { textKey: 'game.notifications.tSpinSingle', tone: 'tspin' }
  return { textKey: 'game.notifications.tSpinNoLines', tone: 'tspin' }
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
      text: i18n.global.t(draft.textKey, draft.textParams ?? {}),
      textKey: draft.textKey,
      textParams: draft.textParams,
      tone: draft.tone,
      expiresAt: now + NOTIFICATION_TTL_MS,
    })
  }

  if (breakdown.reason === 'tSpinNoLines') {
    push({ textKey: 'game.notifications.tSpinNoLines', tone: 'tspin' })
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
    push({ textKey: 'game.notifications.backToBack', tone: 'backToBack' })
  }

  return out
}

export function notificationFromLevelUp(level: number, now: number, id: string): ActionNotification {
  return {
    id,
    text: i18n.global.t('game.notifications.levelUp', { level }),
    textKey: 'game.notifications.levelUp',
    textParams: { level },
    tone: 'level',
    expiresAt: now + NOTIFICATION_TTL_MS,
  }
}
