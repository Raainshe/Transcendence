export const GAME_VARIATIONS = ['marathon', 'sprint', 'ultra', 'multiplayer'] as const
export type GameVariation = (typeof GAME_VARIATIONS)[number]

export const GAME_VARIATION_LABELS: Record<GameVariation, string> = {
  marathon: 'MARATHON',
  sprint: 'SPRINT',
  ultra: 'ULTRA',
  multiplayer: 'MULTIPLAYER',
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

export const GAME_VARIATION_INFO: Record<GameVariation, GameVariationInfo> = {
  marathon: {
    label: GAME_VARIATION_LABELS.marathon,
    tooltip: 'Endless play; level up every 10 lines.',
    description:
      'Classic endless Tetris. Speed increases as you level up every 10 lines cleared. Play until you top out, lock out, or block out.',
    howToWin: 'No win condition — survive as long as possible and chase a high score.',
  },
  sprint: {
    label: GAME_VARIATION_LABELS.sprint,
    tooltip: `Clear ${SPRINT_LINE_GOAL} lines as fast as you can.`,
    description: `A race against the clock. Clear exactly ${SPRINT_LINE_GOAL} lines as quickly as possible. Level stays at 1 for consistent speed.`,
    howToWin: `Clear ${SPRINT_LINE_GOAL} lines before topping out.`,
  },
  ultra: {
    label: GAME_VARIATION_LABELS.ultra,
    tooltip: `Score as much as you can in ${ULTRA_DURATION_MIN} minutes.`,
    description: `Two minutes on the clock. Score as many points as you can before time runs out. Lines still level you up during the match.`,
    howToWin: `Get the highest score possible when the ${ULTRA_DURATION_MIN}:00 timer hits zero.`,
  },
  multiplayer: {
    label: GAME_VARIATION_LABELS.multiplayer,
    tooltip: 'Local multiplayer coming soon; plays as Marathon for now.',
    description:
      'Multiplayer mode is not fully implemented yet. For now this plays like Marathon with a single board.',
    howToWin: 'Same as Marathon until true local multiplayer arrives.',
  },
}

export function getGameVariationInfo(variation: GameVariation): GameVariationInfo {
  return GAME_VARIATION_INFO[variation]
}
