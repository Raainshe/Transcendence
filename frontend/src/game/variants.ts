/**
 * Per-variant engine configuration (slice 8).
 * Pure module — no Vue/Pinia imports.
 */

import type { EngineConfig } from '@/game/types'
import {
  SPRINT_LINE_GOAL,
  ULTRA_DURATION_MS,
  type GameVariation,
} from '@/types/game'

export function engineConfigForVariation(
  variation: GameVariation,
  overrides: Partial<EngineConfig> = {},
): EngineConfig {
  const base: EngineConfig = { variant: variation, ...overrides }

  switch (variation) {
    case 'marathon':
      return { ...base, variant: 'marathon' }
    case 'sprint':
      return {
        ...base,
        variant: 'sprint',
        startLevel: 1,
        sprintLineGoal: overrides.sprintLineGoal ?? SPRINT_LINE_GOAL,
      }
    case 'ultra':
      return {
        ...base,
        variant: 'ultra',
        ultraDurationMs: overrides.ultraDurationMs ?? ULTRA_DURATION_MS,
      }
    case 'multiplayer':
      return { ...base, variant: 'multiplayer' }
    default:
      return base
  }
}
