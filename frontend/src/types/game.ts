import { i18n } from '@/i18n'

export const GAME_VARIATIONS = ['marathon', 'sprint', 'ultra', 'multiplayer'] as const
export type GameVariation = (typeof GAME_VARIATIONS)[number]

export const GAME_VARIATION_LABELS: Record<GameVariation, string> = {
  get marathon() {
    return i18n.global.t('variations.marathon.label')
  },
  get sprint() {
    return i18n.global.t('variations.sprint.label')
  },
  get ultra() {
    return i18n.global.t('variations.ultra.label')
  },
  get multiplayer() {
    return i18n.global.t('variations.multiplayer.label')
  },
}

/** Sprint (40 Lines): total lines to clear to finish. */
export const SPRINT_LINE_GOAL = 40

/** Ultra: match duration in milliseconds (2 minutes). */
export const ULTRA_DURATION_MS = 120_000

export const PLAYER_COUNTS = [1, 2, 3, 4] as const
export type PlayerCount = (typeof PLAYER_COUNTS)[number]

const ULTRA_DURATION_MIN = ULTRA_DURATION_MS / 60_000

export type GameVariationInfo = {
  label: string
  tooltip: string
  description: string
  howToWin: string
}

function createVariationInfo(variation: GameVariation): GameVariationInfo {
  return {
    get label() {
      return i18n.global.t(`variations.${variation}.label`)
    },
    get tooltip() {
      return i18n.global.t(`variations.${variation}.tooltip`, {
        goal: SPRINT_LINE_GOAL,
        minutes: ULTRA_DURATION_MIN,
      })
    },
    get description() {
      return i18n.global.t(`variations.${variation}.description`, {
        goal: SPRINT_LINE_GOAL,
        minutes: ULTRA_DURATION_MIN,
      })
    },
    get howToWin() {
      return i18n.global.t(`variations.${variation}.howToWin`, {
        goal: SPRINT_LINE_GOAL,
        minutes: ULTRA_DURATION_MIN,
      })
    },
  }
}

export const GAME_VARIATION_INFO: Record<GameVariation, GameVariationInfo> = {
  marathon: createVariationInfo('marathon'),
  sprint: createVariationInfo('sprint'),
  ultra: createVariationInfo('ultra'),
  multiplayer: createVariationInfo('multiplayer'),
}

export function getGameVariationInfo(variation: GameVariation): GameVariationInfo {
  return GAME_VARIATION_INFO[variation]
}
