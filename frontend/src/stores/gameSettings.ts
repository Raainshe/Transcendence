import { defineStore } from 'pinia'
import { ref } from 'vue'

import type { GameVariation, PlayerCount } from '@/types/game'

// TODO: When implementing the game, couple `playerCount` and `variation` so that
// `sprint` forces 1 player and `multiplayer` forces 2+. For now they remain
// independent so the menu's two cyclers stay simple and predictable.
export const useGameSettingsStore = defineStore('gameSettings', () => {
  const variation = ref<GameVariation>('sprint')
  const playerCount = ref<PlayerCount>(1)

  return { variation, playerCount }
})
