export const GAME_VARIATIONS = ['sprint', 'multiplayer'] as const
export type GameVariation = (typeof GAME_VARIATIONS)[number]

export const GAME_VARIATION_LABELS: Record<GameVariation, string> = {
  sprint: 'SPRINT',
  multiplayer: 'MULTIPLAYER',
}

export const PLAYER_COUNTS = [1, 2, 3, 4] as const
export type PlayerCount = (typeof PLAYER_COUNTS)[number]
